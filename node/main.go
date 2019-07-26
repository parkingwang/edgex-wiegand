package main

import (
	"flag"
	"fmt"
	"github.com/nextabc-lab/edgex-go"
	"github.com/nextabc-lab/edgex-wiegand"
	"github.com/tidwall/evio"
	"github.com/yoojia/go-at"
	"github.com/yoojia/go-value"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

const (
	evioUdpProto = "udp-net"
)

func main() {
	edgex.Run(wiegrandApp)
}

func wiegrandApp(ctx edgex.Context) error {
	fileName := flag.String("c", edgex.DefaultConfName, "config file name")
	config := ctx.LoadConfigByName(*fileName)
	// Init
	log := ctx.Log()
	ctx.InitialWithConfig(config)

	rpcAddress := value.Of(config["RpcAddress"]).String()
	eventTopic := value.Of(config["Topic"]).String()

	// 主板参数
	boardOpts := value.Of(config["BoardOptions"]).MustMap()
	serialNumber := uint32(value.Of(boardOpts["serialNumber"]).MustInt64())
	doorCount := int(value.Of(boardOpts["doorCount"]).Int64OrDefault(4))

	trigger := ctx.NewTrigger(edgex.TriggerOptions{
		Topic:           eventTopic,
		AutoInspectFunc: wiegand.FuncTriggerNode(serialNumber, doorCount),
	})

	// AT指令解析
	atRegistry := at.NewAtRegister()
	wiegand.AtCommands(atRegistry, serialNumber)

	// Endpoint 远程控制服务
	// 使用UDP客户端连接
	udpOpts := value.Of(config["UdpClientOptions"]).MustMap()
	remoteAddress := value.Of(udpOpts["remoteAddress"]).String()
	log.Debugf("连接UDP主板地址: %s://%s", evioUdpProto, remoteAddress)
	conn, err := makeUdpConn(remoteAddress)
	if nil != err {
		return err
	}
	endpoint := ctx.NewEndpoint(edgex.EndpointOptions{
		RpcAddr:         rpcAddress,
		SerialExecuting: true, // 微耕品牌设置不支持并发处理
		AutoInspectFunc: wiegand.FuncEndpointNode(serialNumber, doorCount),
	})
	endpoint.Serve(wiegand.FuncEndpointHandler(ctx, endpoint, atRegistry, conn))

	// Trigger 事件监听服务
	// 使用Socket服务端接收消息
	var server evio.Events
	server.NumLoops = 1
	server.Data = wiegand.FuncTriggerHandler(ctx, trigger, serialNumber)
	opts := value.Of(config["UdpServerOptions"]).MustMap()
	address := fmt.Sprintf("%s://%s", evioUdpProto, value.Of(opts["listenAddress"]).String())
	log.Debug("开启UDP服务监听: ", address)
	defer log.Debug("停止UDP服务端")

	// 启用Trigger服务
	trigger.Startup()
	defer trigger.Shutdown()
	// 启动Endpoint服务
	endpoint.Startup()
	defer endpoint.Shutdown()

	return evio.Serve(server, address)
}
