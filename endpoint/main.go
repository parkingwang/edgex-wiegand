package main

import (
	"encoding/hex"
	"fmt"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/nextabc-lab/edgex-go"
	"github.com/yoojia/go-at"
	"github.com/yoojia/go-value"
	"runtime"
	"time"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com

func main() {
	edgex.Run(endpoint)
}

func endpoint(ctx edgex.Context) error {
	config := ctx.LoadConfig()
	deviceName := value.Of(config["Name"]).String()
	rpcAddress := value.Of(config["RpcAddress"]).String()

	sockOpts := value.Of(config["SocketClientOptions"]).MustMap()
	remoteAddress := value.Of(sockOpts["remoteAddress"]).String()
	readTimeout := value.Of(sockOpts["readTimeout"]).DurationOfDefault(time.Second)
	writeTimeout := value.Of(sockOpts["writeTimeout"]).DurationOfDefault(time.Second)

	boardOpts := value.Of(config["BoardOptions"]).MustMap()
	serialNumber := uint32(value.Of(boardOpts["serialNumber"]).MustInt64())
	doorCount := value.Of(boardOpts["doorCount"]).Int64OrDefault(4)

	// AT指令解析
	atRegistry := at.NewAtRegister()
	atCommands(atRegistry, serialNumber)

	ctx.Log().Debugf("连接目标地址: [udp://%s]", remoteAddress)
	conn, err := makeUdpConn(remoteAddress)
	if nil != err {
		return err
	}

	buffer := make([]byte, 64)
	endpoint := ctx.NewEndpoint(edgex.EndpointOptions{
		Name:        deviceName,
		RpcAddr:     rpcAddress,
		InspectFunc: inspectFunc(serialNumber, int(doorCount)),
	})

	// 处理控制指令
	endpoint.Serve(func(msg edgex.Message) (out edgex.Message) {
		atCmd := string(msg.Body())
		ctx.Log().Debug("接收到控制指令: " + atCmd)
		cmd, err := atRegistry.Apply(atCmd)
		if nil != err {
			return edgex.NewMessageString(deviceName, "EX=ERR:"+err.Error())
		}
		ctx.Log().Debug("东控指令码: " + hex.EncodeToString(cmd))
		// Write
		if err := tryWrite(conn, cmd, writeTimeout); nil != err {
			return edgex.NewMessageString(deviceName, "EX=ERR:"+err.Error())
		}
		// Read
		var n = int(0)
		for i := 0; i < 2; i++ {
			if n, err = tryRead(conn, buffer, readTimeout); nil != err {
				ctx.Log().Errorf("读取设备响应数据出错[%d]: %s", i, err.Error())
				<-time.After(time.Millisecond * 500)
			} else {
				break
			}
		}
		// parse
		if n > 0 {
			if outCmd, err := dongk.ParseCommand(buffer); nil != err {
				return edgex.NewMessageString(deviceName, "EX=ERR:"+err.Error())
			} else if outCmd.Success() {
				return edgex.NewMessageString(deviceName, fmt.Sprintf("EX=OK"))
			} else {
				return edgex.NewMessageString(deviceName, "EX=ERR:NOT_OK")
			}
		} else {
			return edgex.NewMessageString(deviceName, "EX=ERR:NO_REPLY")
		}

	})

	endpoint.Startup()
	defer endpoint.Shutdown()

	return ctx.TermAwait()
}

func inspectFunc(sn uint32, doorCount int) func() edgex.Inspect {
	deviceOf := func(doorId int) edgex.Device {
		return edgex.Device{
			Name:    fmt.Sprintf("ENDPOINT-%d-%d", sn, doorId),
			Desc:    fmt.Sprintf("%d号门-控制开关", doorId),
			Type:    edgex.DeviceTypeEndpoint,
			Virtual: true,
			Address: "grpc://ep-dk.edgex.io:5570",
			Command: fmt.Sprintf("AT+OPEN=%d", doorId),
		}
	}
	return func() edgex.Inspect {
		devices := make([]edgex.Device, doorCount)
		for d := 0; d < doorCount; d++ {
			devices[d] = deviceOf(d + 1)
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
