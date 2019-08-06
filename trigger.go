package wiegand

import (
	"encoding/hex"
	"fmt"
	"github.com/nextabc-lab/edgex-go"
	"github.com/tidwall/evio"
	"go.uber.org/zap"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

const (
	// 设备地址格式：　READER - 序列号 - 门号 - 方向
	// 门禁设备，一个门对应两个输入端
	readerVirtualIdFormat = "READER-%d-%d-%s"
)

// 创建Trigger处理函数
func FuncTriggerHandler(ctx edgex.Context, trigger edgex.Trigger, serialNumber uint32) func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
	log := ctx.Log()
	return func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		cmd, err := ParseCommand(in)
		if nil != err {
			log.Debugf("接收到非微耕数据格式数据: ERR= %s, DATA= %v", err.Error(), in)
			return []byte("EX=ERR:INVALID_DK_COMMAND"), action
		}
		// 非监控数据，忽略
		if cmd.FuncId != FunIdBoardState {
			log.Debugf("接收到非监控状态数据: FunId= %x", cmd.FuncId)
			return []byte("EX=ERR:INVALID_DK_STATE"), action
		}
		// 只接收允许的序列号的数据
		// TODO 运行多序列号同时接入同一个节点
		if cmd.SerialNum != serialNumber {
			log.Debugf("接收到未知序列号数据: SN= %d", cmd.SerialNum)
			return []byte("EX=ERR:UNKNOWN_BOARD_SN"), action
		}
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("微耕控制指令码: ", hex.EncodeToString(in))
		})
		// 控制指令数据：
		bytes, card, doorId, direct, rType := EventToCard(cmd)
		// 最后执行控制指令：刷卡数据
		virtualNodeId := fmt.Sprintf(readerVirtualIdFormat, cmd.SerialNum, doorId, DirectName(direct))
		log.Debugf("接收到刷卡数据, Device: %s, Card: %s, Type: %s", virtualNodeId, card, TypeName(rType))
		if rType != 1 {
			log.Debug("接收到非刷卡类型数据")
			return []byte("EX=ERR:IGNORE_RECORD_TYPE"), action
		}
		if err := trigger.PublishEvent(virtualNodeId, bytes); nil != err {
			log.Error("触发事件出错: ", err)
			return []byte("EX=ERR:" + err.Error()), action
		} else {
			return []byte("EX=OK:WG_EVENT"), action
		}
	}
}

// 创建Trigger节点消息函数
func FuncTriggerNode(serialNum uint32, doorCount int) func() edgex.MainNodeInfo {
	deviceOf := func(doorId, direct int) *edgex.VirtualNodeInfo {
		directName := DirectName(byte(direct))
		return &edgex.VirtualNodeInfo{
			VirtualId: fmt.Sprintf(readerVirtualIdFormat, serialNum, doorId, directName),
			MajorId:   fmt.Sprintf("%d:%d", serialNum, doorId),
			MinorId:   directName,
			Desc:      fmt.Sprintf("%d号门-%s-读卡器", doorId, directName),
			Virtual:   true,
		}
	}
	return func() edgex.MainNodeInfo {
		nodes := make([]*edgex.VirtualNodeInfo, doorCount*2)
		for d := 0; d < doorCount; d++ {
			nodes[d*2] = deviceOf(d+1, DirectIn)
			nodes[d*2+1] = deviceOf(d+1, DirectOut)
		}
		return edgex.MainNodeInfo{
			NodeType:     edgex.NodeTypeTrigger,
			Vendor:       VendorName,
			ConnDriver:   ConnectionDriver,
			VirtualNodes: nodes,
		}
	}
}
