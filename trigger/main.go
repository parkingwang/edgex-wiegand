package main

import (
	"encoding/hex"
	"fmt"
	"github.com/nextabc-lab/edgex-go"
	"github.com/nextabc-lab/edgex-wiegand"
	"github.com/tidwall/evio"
	"github.com/yoojia/go-value"
	"go.uber.org/zap"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

const (
	// 设备地址格式：　READER - 序列号 - 门号 - 方向
	// 门禁设备，一个门对应两个输入端
	virtualDeviceName = "READER-%d-%d-%s"
)

func main() {
	edgex.Run(trigger)
}

func trigger(ctx edgex.Context) error {
	config := ctx.LoadConfig()
	nodeName := value.Of(config["NodeName"]).String()
	eventTopic := value.Of(config["Topic"]).String()

	boardOpts := value.Of(config["BoardOptions"]).MustMap()
	serialNumber := uint32(value.Of(boardOpts["serialNumber"]).MustInt64())
	doorCount := value.Of(boardOpts["doorCount"]).Int64OrDefault(4)

	trigger := ctx.NewTrigger(edgex.TriggerOptions{
		NodeName:        nodeName,
		Topic:           eventTopic,
		InspectNodeFunc: nodeFunc(nodeName, serialNumber, int(doorCount)),
	})

	var server evio.Events

	opts := value.Of(config["SocketServerOptions"]).MustMap()
	server.NumLoops = 1

	log := ctx.Log()

	server.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		cmd, err := wiegand.ParseCommand(in)
		if nil != err {
			log.Errorw("接收到非微耕数据格式数据", "error", err, "data", in)
			return []byte("EX=ERR:INVALID_DK_COMMAND"), action
		}
		// 非监控数据，忽略
		if cmd.FuncId != wiegand.FunIdBoardState {
			log.Debug("接收到非监控状态数据")
			return []byte("EX=ERR:INVALID_DK_STATE"), action
		}
		if cmd.SerialNum != serialNumber {
			log.Debug("接收到未知序列号数据")
			return []byte("EX=ERR:UNKNOWN_BOARD_SN"), action
		}
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("微耕控制指令码: ", hex.EncodeToString(in))
		})
		// 控制指令数据：
		bytes, card, doorId, direct, rType := cmdToJSON(cmd)
		// 最后执行控制指令：刷卡数据
		deviceName := fmt.Sprintf(virtualDeviceName, cmd.SerialNum, doorId, wiegand.DirectName(direct))
		log.Debugf("接收到刷卡数据, Device: %s, Card: %s, Type: %s", deviceName, card, wiegand.TypeName(rType))
		if rType != 1 {
			log.Debug("接收到非刷卡类型数据")
			return []byte("EX=ERR:IGNORE_RECORD_TYPE"), action
		}
		if err := trigger.SendEventMessage(deviceName, bytes); nil != err {
			log.Error("触发事件出错: ", err)
			return []byte("EX=ERR:" + err.Error()), action
		} else {
			return []byte("EX=OK:DK_EVENT"), action
		}
	}

	server.Opened = func(c evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		log.Debug("接受客户端: ", c.RemoteAddr())
		return
	}

	server.Closed = func(c evio.Conn, err error) (action evio.Action) {
		log.Debug("断开客户端: ", c.RemoteAddr())
		return
	}

	address := value.Of(opts["address"]).MustStringArray()
	log.Debug("开启Evio服务端: ", address)
	defer log.Debug("停止Evio服务端")

	// 启用Trigger服务
	trigger.Startup()
	defer trigger.Shutdown()

	return evio.Serve(server, address...)
}

func nodeFunc(nodeName string, serialNum uint32, doorCount int) func() edgex.MainNode {
	deviceOf := func(doorId, direct int) edgex.VirtualNode {
		directName := wiegand.DirectName(byte(direct))
		return edgex.VirtualNode{
			Major:   fmt.Sprintf("%d:%d", serialNum, doorId),
			Minor:   directName,
			Desc:    fmt.Sprintf("%d号门-%s-读卡器", doorId, directName),
			Virtual: true,
		}
	}
	return func() edgex.MainNode {
		nodes := make([]edgex.VirtualNode, doorCount*2)
		for d := 0; d < doorCount; d++ {
			nodes[d*2] = deviceOf(d+1, wiegand.DirectIn)
			nodes[d*2+1] = deviceOf(d+1, wiegand.DirectOut)
		}
		return edgex.MainNode{
			NodeType:     edgex.NodeTypeTrigger,
			NodeName:     nodeName,
			Vendor:       wiegand.VendorName,
			ConnDriver:   wiegand.ConnectionDriver,
			VirtualNodes: nodes,
		}
	}
}
