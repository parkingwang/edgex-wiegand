package main

import (
	"encoding/hex"
	"fmt"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/nextabc-lab/edgex-go"
	"github.com/nextabc-lab/edgex-wiegand"
	"github.com/yoojia/go-at"
	"github.com/yoojia/go-value"
	"go.uber.org/zap"
	"time"
)

//
// Author: 陈哈哈 bitschen@163.com

func main() {
	edgex.Run(endpoint)
}

func endpoint(ctx edgex.Context) error {
	config := ctx.LoadConfig()
	nodeName := value.Of(config["NodeName"]).String()
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

	log := ctx.Log()
	log.Debugf("连接目标地址: [udp://%s]", remoteAddress)
	conn, err := makeUdpConn(remoteAddress)
	if nil != err {
		return err
	}

	buffer := make([]byte, 64)
	endpoint := ctx.NewEndpoint(edgex.EndpointOptions{
		NodeName:        nodeName,
		RpcAddr:         rpcAddress,
		SerialExecuting: true, // 微耕品牌设置不支持并发处理
		InspectFunc:     inspectFunc(serialNumber, int(doorCount)),
	})

	// 处理控制指令
	endpoint.Serve(func(msg edgex.Message) (out edgex.Message) {
		atCmd := string(msg.Body())

		log.Debug("接收到控制指令: " + atCmd)
		cmd, err := atRegistry.Apply(atCmd)
		if nil != err {
			return endpoint.NextMessage(nodeName, []byte("EX=ERR:BAD_CMD:"+err.Error()))
		}
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("微耕指令码: " + hex.EncodeToString(cmd))
		})
		// Write
		if err := tryWrite(conn, cmd, writeTimeout); nil != err {
			return endpoint.NextMessage(nodeName, []byte("EX=ERR:WRITE:"+err.Error()))
		}
		// Read
		var n = int(0)
		for i := 0; i < 2; i++ {
			if n, err = tryRead(conn, buffer, readTimeout); nil != err {
				log.Errorf("读取设备响应数据出错[%d]: %s", i, err.Error())
				<-time.After(time.Millisecond * 500)
			} else {
				break
			}
		}
		// parse
		reply := "EX=ERR:NO_REPLY"
		if n > 0 {
			if outCmd, err := wiegand.ParseCommand(buffer); nil != err {
				log.Error("解析响应数据出错", err)
				reply = "EX=ERR:PARSE_ERR"
			} else if outCmd.Success() {
				reply = "EX=OK"
			} else {
				reply = "EX=ERR:NOT_OK"
			}
		}
		log.Debug("接收到控制响应: " + reply)
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("响应码: " + hex.EncodeToString(buffer))
		})
		return endpoint.NextMessage(nodeName, []byte(reply))
	})

	endpoint.Startup()
	defer endpoint.Shutdown()

	return ctx.TermAwait()
}

func inspectFunc(sn uint32, doorCount int) func() edgex.Inspect {
	deviceOf := func(doorId int) edgex.VirtualNode {
		// Address 可以自动从环境变量中获取
		return edgex.VirtualNode{
			VirtualNodeName: fmt.Sprintf("SWITCH-%d-%d", sn, doorId),
			Desc:            fmt.Sprintf("%d号门-控制开关", doorId),
			Type:            edgex.NodeTypeEndpoint,
			Virtual:         true,
			Command:         fmt.Sprintf("AT+OPEN=%d", doorId),
		}
	}
	return func() edgex.Inspect {
		nodes := make([]edgex.VirtualNode, doorCount)
		for d := 0; d < doorCount; d++ {
			nodes[d] = deviceOf(d + 1)
		}
		return edgex.Inspect{
			Vendor:       wiegand.VendorName,
			DriverName:   wiegand.DriverName,
			VirtualNodes: nodes,
		}
	}
}
