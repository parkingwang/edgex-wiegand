package main

import (
	"fmt"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/nextabc-lab/edgex-go"
	"github.com/tidwall/evio"
	"github.com/yoojia/go-value"
	"runtime"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

const (
	// 设备地址格式：　TRIGGER/SN/DOOR_ID/DIRECT
	deviceAddr = "TRIGGER/%d/%d/%s"
)

func main() {
	edgex.Run(trigger)
}

func trigger(ctx edgex.Context) error {
	config := ctx.LoadConfig()
	triggerName := value.Of(config["Name"]).String()
	eventTopic := value.Of(config["Topic"]).String()

	boardOpts := value.Of(config["BoardOptions"]).MustMap()
	serialNumber := uint32(value.Of(boardOpts["serialNumber"]).MustInt64())
	doorCount := value.Of(boardOpts["doorCount"]).Int64OrDefault(4)

	trigger := ctx.NewTrigger(edgex.TriggerOptions{
		Name:        triggerName,
		Topic:       eventTopic,
		InspectFunc: inspectFunc(serialNumber, int(doorCount), eventTopic),
	})

	var server evio.Events

	opts := value.Of(config["SocketServerOptions"]).MustMap()
	server.NumLoops = 1

	server.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		cmd, err := dongk.ParseCommand(in)
		if nil != err {
			ctx.Log().Errorw("接收到非东控数据格式数据", "error", err, "data", in)
			return []byte("EX=ERR:INVALID_CMD_SIZE"), action
		}
		// 非监控数据，忽略
		if cmd.FuncId != dongk.FunIdBoardState {
			ctx.Log().Debug("接收到非监控状态数据")
			return []byte("EX=ERR:INVALID_STATE"), action
		}
		if cmd.SerialNum != serialNumber {
			ctx.Log().Debug("接收到未知序列号数据")
			return []byte("EX=ERR:UNKNOWN_SN"), action
		}
		// 控制指令数据：
		bytes, doorId, direct := cmdToJSON(cmd)
		// 最后执行控制指令：刷卡数据
		// 地址： TRIGGER/序列号/门号/方向
		deviceName := fmt.Sprintf(deviceAddr, cmd.SerialNum, doorId, dongk.DirectName(direct))
		if err := trigger.SendEventMessage(edgex.NewMessage([]byte(deviceName), bytes)); nil != err {
			ctx.Log().Error("触发事件出错: ", err)
			return []byte("EX=ERR:" + err.Error()), action
		} else {
			return []byte("EX=OK:DK_EVENT"), action
		}
	}

	server.Opened = func(c evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		ctx.Log().Debug("接受客户端: ", c.RemoteAddr())
		return
	}

	server.Closed = func(c evio.Conn, err error) (action evio.Action) {
		ctx.Log().Debug("断开客户端: ", c.RemoteAddr())
		return
	}

	// 启用Trigger服务
	trigger.Startup()
	defer trigger.Shutdown()

	address := value.Of(opts["address"]).MustStringArray()
	ctx.Log().Debug("开启Evio服务端: ", address)
	defer ctx.Log().Debug("停止Evio服务端")
	return evio.Serve(server, address...)
}

func inspectFunc(sn uint32, doorCount int, eventTopic string) func() edgex.Inspect {
	deviceOf := func(doorId, direct int) edgex.Device {
		directName := dongk.DirectName(byte(direct))
		return edgex.Device{
			Name:       fmt.Sprintf(deviceAddr, sn, doorId, directName),
			Desc:       fmt.Sprintf("%d号门%s读卡器", doorId, directName),
			Type:       edgex.DeviceTypeTrigger,
			Virtual:    true,
			EventTopic: eventTopic,
		}
	}
	return func() edgex.Inspect {
		devices := make([]edgex.Device, doorCount*2)
		for d := 0; d < doorCount; d++ {
			devices[d*2] = deviceOf(d+1, dongk.DirectIn)
			devices[d*2+1] = deviceOf(d+1, dongk.DirectOut)
		}
		return edgex.Inspect{
			HostOS:     runtime.GOOS,
			HostArch:   runtime.GOARCH,
			Vendor:     dongk.VendorName,
			DriverName: dongk.DriverName,
			Devices:    devices,
		}
	}
}
