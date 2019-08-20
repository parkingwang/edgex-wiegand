package wiegand

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/nextabc-lab/edgex-go"
	"github.com/nextabc-lab/edgex-go/extra"
	"github.com/parkingwang/go-wg26"
	"github.com/tidwall/evio"
	"go.uber.org/zap"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

// 创建Trigger处理函数
func FuncTriggerHandler(ctx edgex.Context, trigger edgex.Trigger, serialNumber uint32) func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
	log := ctx.Log()
	return func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		cmd, err := ParseCommand(in)
		if nil != err {
			log.Debugf("接收到非微耕数据格式数据: ERR= %s, DATA= %v", err.Error(), in)
			return []byte("EX=ERR:INVALID_DK_DATA"), action
		}
		// TODO 其它事件：1. 按钮开门事件；2. 报警事件；
		// 非监控数据，忽略
		if cmd.FuncId != FunIdBoardState {
			log.Debugf("只处理监控事件，忽略: FunId= %x", cmd.FuncId)
			return []byte("EX=ERR:INVALID_DK_STATE"), action
		}
		// 只接收允许的序列号的数据
		// TODO 运行多序列号同时接入同一个节点
		if cmd.SerialNum != serialNumber {
			log.Debugf("接收到未知序列号数据: SN= %d", cmd.SerialNum)
			return []byte("EX=ERR:UNKNOWN_BOARD_SN"), action
		}
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("微耕发送事件码: ", hex.EncodeToString(in))
		})
		// 控制指令数据：
		event := parseCardEvent(cmd)
		log.Debugf("接收到控制器事件, DoorId: %d, Card: %s, EventType: %s", event.DoorId, event.CardNO, event.Type)
		if event.Type != extra.TypeCard {
			log.Debug("只处理刷卡类型事件，忽略")
			return []byte("EX=ERR:UNSUPPORTED_TYPE"), action
		}
		data, err := json.Marshal(event)
		if nil != err {
			log.Error("JSON序列化错误", err)
			return []byte("EX=ERR:JSON_ERROR"), action
		}
		if err := trigger.PublishEvent(
			makeGroupId(event.BoardId),
			makeMajorId(int(event.DoorId)),
			directName(event.Direct),
			data,
			trigger.GenerateEventId()); nil != err {
			log.Error("触发事件出错: ", err)
			return []byte("EX=ERR:" + err.Error()), action
		} else {
			return []byte("EX=OK:ACCEPTED"), action
		}
	}
}

func parseCardEvent(cmd *Command) extra.CardEvent {
	reader := cmd.DataReader()
	idx := reader.NextUint32()
	aType := reader.NextByte()
	state := reader.NextByte()
	doorId := reader.NextByte()
	direct := reader.NextByte()
	card := reader.NextUint32()
	return extra.CardEvent{
		SerialNum: cmd.SerialNum,
		BoardId:   cmd.SerialNum,
		DoorId:    doorId,
		Direct:    directVal(direct),
		CardNO:    wg26.ParseFromCardNumber(fmt.Sprintf("%d", card)).CardSN,
		Type:      typeName(aType),
		State:     stateName(state),
		Index:     idx,
	}
}

// 创建Trigger节点消息函数
func FuncTriggerProperties(serialNum uint32, doorCount int) func() edgex.MainNodeProperties {
	deviceOf := func(doorId int, direct string) *edgex.VirtualNodeProperties {
		return &edgex.VirtualNodeProperties{
			GroupId:     makeGroupId(serialNum),
			MajorId:     makeMajorId(doorId),
			MinorId:     direct,
			Description: fmt.Sprintf("控制器#%d-%d号门-%s", serialNum, doorId, direct),
			Virtual:     true,
		}
	}
	return func() edgex.MainNodeProperties {
		nodes := make([]*edgex.VirtualNodeProperties, doorCount*2)
		for d := 0; d < doorCount; d++ {
			nodes[d*2] = deviceOf(d+1, "IN")
			nodes[d*2+1] = deviceOf(d+1, "OUT")
		}
		return edgex.MainNodeProperties{
			NodeType:     edgex.NodeTypeTrigger,
			Vendor:       VendorName,
			ConnDriver:   ConnectionDriver,
			VirtualNodes: nodes,
		}
	}
}
