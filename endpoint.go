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
	ioTimeout = time.Second * 2
)

func FuncEndpointHandler(ctx edgex.Context, atRegistry *at.Registry, conn *net.UDPConn) func(msg edgex.Message) (out []byte) {
	log := ctx.Log()
	buffer := make([]byte, 64)
	return func(msg edgex.Message) []byte {
		atCmd := string(msg.Body())
		log.Debug("接收到AT控制指令: " + atCmd)
		wgCmd, err := atRegistry.Transformer(atCmd)
		if nil != err {
			return []byte("EX=ERR:UNKNOWN_CMD:" + err.Error())
		}
		ctx.LogIfVerbose(func(log *zap.SugaredLogger) {
			log.Debug("发送指令码: " + hex.EncodeToString(wgCmd.Payload))
		})
		// Write
		if err := tryWrite(conn, wgCmd.Payload, ioTimeout); nil != err {
			return []byte("EX=ERR:SEND:" + err.Error())
		}
		// Read
		read, err := tryRead(conn, buffer, ioTimeout)
		if nil != err {
			log.Errorf("读取设备响应数据出错: " + err.Error())
			return []byte("EX=OK:SEND_OK/RECV_ERR:" + err.Error())
		}
		// 如果是[Open]等指令，共享EventId
		attrKey := attrKeyRpcEventId(msg.BoardId(), msg.MajorId())
		if "OPEN" == wgCmd.Name {
			log.Debugf("共享RPC.EventId: %d", msg.EventId())
			ctx.StoreAttr(attrKey, msg.EventId())
		} else {
			ctx.RemoveAttr(attrKey)
		}
		// parse out
		data := buffer[:read]
		if reply, err := ParseCommand(data); nil != err {
			log.Error("解析响应数据出错", err)
			return []byte("EX=ERR:PARSE_ERR:" + err.Error())
		} else if reply.Success() {
			return []byte("EX=OK:SUCCESS")
		} else {
			return []byte("EX=ERR:FAILED")
		}
	}
}

func FuncEndpointProperties(serialNum uint32, doorCount int) func() edgex.MainNodeProperties {
	deviceOf := func(doorId int) *edgex.VirtualNodeProperties {
		cmd := fmt.Sprintf("AT+OPEN=%d", doorId)
		return &edgex.VirtualNodeProperties{
			BoardId:     makeBoardId(serialNum),
			MajorId:     makeMajorId(doorId),
			MinorId:     "SW",
			DeviceType:  "switch",
			Description: fmt.Sprintf("微耕#%d/%d号门/开关", serialNum, doorId),
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
