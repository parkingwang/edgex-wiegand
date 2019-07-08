package main

import (
	"encoding/hex"
	"fmt"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/nextabc-lab/edgex-go"
	"github.com/tidwall/evio"
	"github.com/yoojia/go-value"
	"go.uber.org/zap"
)

//
// Author: 陈哈哈 bitschen@163.com
//

const (
	// 设备地址格式：　READER - 序列号 - 门号 - 方向
	// 东控门禁设备，一个门对应两个输入端
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
		NodeName:    nodeName,
		Topic:       eventTopic,
		InspectFunc: inspectFunc(serialNumber, int(doorCount), eventTopic),
	})

	var server evio.Events

	opts := value.Of(config["SocketServerOptions"]).MustMap()
	server.NumLoops = 1

	log := ctx.Log()

	server.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		cmd, err := dongk.ParseCommand(in)
		if nil != err {
			log.Errorw("接收到非东控数据格式数据", "error", err, "data", in)
			return []byte("EX=ERR:INVALID_DK_COMMAND"), action
		}
		// 非监控数据，忽略
		if cmd.FuncId != dongk.FunIdBoardState {
			log.Debug("接收到非监控状态数据")
			return []byte("EX=ERR:INVALID_DK_STATE"), action
		}
		if cmd.SerialNum != serialNumber {
			log.Debug("接收到未知序列号数据")
			return []byte("EX=ERR:UNKNOWN_BOARD_SN"), action
		}
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("东控制指令码: ", hex.EncodeToString(in))
		})
		// 控制指令数据：
		bytes, card, doorId, direct, rType := cmdToJSON(cmd)
		// 最后执行控制指令：刷卡数据
		deviceName := fmt.Sprintf(virtualDeviceName, cmd.SerialNum, doorId, dongk.DirectName(direct))
		log.Debugf("接收到刷卡数据, Device: %s, Card: %s, Type: %s", deviceName, card, dongk.TypeName(rType))
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

func inspectFunc(sn uint32, doorCount int, eventTopic string) func() edgex.Inspect {
	deviceOf := func(doorId, direct int) edgex.VirtualNode {
		directName := dongk.DirectName(byte(direct))
		return edgex.VirtualNode{
			VirtualNodeName: fmt.Sprintf(virtualDeviceName, sn, doorId, directName),
			Desc:            fmt.Sprintf("%d号门-%s-读卡器", doorId, directName),
			Type:            edgex.NodeTypeTrigger,
			Virtual:         true,
			EventTopic:      eventTopic,
		}
	}
	return func() edgex.Inspect {
		nodes := make([]edgex.VirtualNode, doorCount*2)
		for d := 0; d < doorCount; d++ {
			nodes[d*2] = deviceOf(d+1, dongk.DirectIn)
			nodes[d*2+1] = deviceOf(d+1, dongk.DirectOut)
		}
		return edgex.Inspect{
			Vendor:       dongk.VendorName,
			DriverName:   dongk.DriverName,
			VirtualNodes: nodes,
		}
	}
}
