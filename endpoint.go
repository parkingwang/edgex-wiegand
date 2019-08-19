package wiegand

import (
	"encoding/hex"
	"fmt"
	"github.com/nextabc-lab/edgex-go"
	"github.com/yoojia/go-at"
	"go.uber.org/zap"
	"net"
	"time"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

const (
	ioTimeout             = time.Second * 3
	switchVirtualIdFormat = "SWITCH-%d-%d"
)

func FuncEndpointHandler(ctx edgex.Context, endpoint edgex.Endpoint, atRegistry *at.AtRegister, conn *net.UDPConn) func(msg edgex.Message) (out []byte) {
	log := ctx.Log()
	buffer := make([]byte, 64)
	return func(msg edgex.Message) (out []byte) {
		atCmd := string(msg.Body())
		eventId := msg.EventId()
		vnId := msg.VirtualNodeId()

		log.Debug("接收到控制指令: " + atCmd)
		vendorCommand, err := atRegistry.Apply(atCmd)
		if nil != err {
			return []byte("EX=ERR:UNKNOWN_CMD:" + err.Error())
		}
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("微耕指令码: " + hex.EncodeToString(vendorCommand))
		})
		// Write
		if err := tryWrite(conn, vendorCommand, ioTimeout); nil != err {
			return []byte("EX=ERR:WRITE:" + err.Error())
		}
		// Read
		var n = int(0)
		for i := 0; i < 2; i++ {
			if n, err = tryRead(conn, buffer, ioTimeout); nil != err {
				log.Errorf("读取设备响应数据出错[%d]: %s", i, err.Error())
				<-time.After(time.Millisecond * 500)
			} else {
				break
			}
		}
		// parse
		reply := "EX=ERR:NO_REPLY"
		success := false
		if n > 0 {
			if outCmd, err := ParseCommand(buffer); nil != err {
				log.Error("解析响应数据出错", err)
				reply = "EX=ERR:PARSE_ERR"
			} else if outCmd.Success() {
				reply = "EX=OK:SUCCESS"
				success = true
			} else {
				reply = "EX=ERR:FAILED"
			}
		}
		log.Debug("接收到控制响应: " + reply)
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("响应码: " + hex.EncodeToString(buffer))
		})
		// 设备被驱动后，发出Action消息广播。用于联动。
		go func() {
			action := "ACT:"
			if success {
				action += "SUCCESS"
			} else {
				action += "FAILED"
			}
			if err := endpoint.PublishActionMessage(
				endpoint.NewMessageOf(vnId, []byte(action), eventId)); nil != err {
				log.Error("发出Action广播出错：", err)
			}
		}()
		return []byte(reply)
	}
}

func FuncEndpointProperties(serialNum uint32, doorCount int) func() edgex.MainNodeProperties {
	deviceOf := func(doorId int) *edgex.VirtualNodeProperties {
		cmd := fmt.Sprintf("AT+OPEN=%d", doorId)
		return &edgex.VirtualNodeProperties{
			VirtualId:   fmt.Sprintf(switchVirtualIdFormat, serialNum, doorId),
			MajorId:     fmt.Sprintf("%d", serialNum),
			MinorId:     fmt.Sprintf("%d", doorId),
			Description: fmt.Sprintf("%d号门-控制开关", doorId),
			Virtual:     true,
			StateCommands: map[string]string{
				"TRIGGER": cmd,
				"OPEN":    cmd,
			},
		}
	}
	return func() edgex.MainNodeProperties {
		nodes := make([]*edgex.VirtualNodeProperties, doorCount)
		for d := 0; d < doorCount; d++ {
			nodes[d] = deviceOf(d + 1)
		}
		return edgex.MainNodeProperties{
			NodeType:     edgex.NodeTypeEndpoint,
			Vendor:       VendorName,
			ConnDriver:   ConnectionDriver,
			VirtualNodes: nodes,
		}
	}
}

func tryWrite(conn *net.UDPConn, bs []byte, to time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(to)); nil != err {
		return err
	}
	if _, err := conn.Write(bs); nil != err {
		return err
	} else {
		return nil
	}
}

func tryRead(conn *net.UDPConn, buffer []byte, to time.Duration) (n int, err error) {
	if err := conn.SetReadDeadline(time.Now().Add(to)); nil != err {
		return 0, err
	}
	return conn.Read(buffer)
}
