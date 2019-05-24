package main

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/nextabc-lab/edgex-go"
	"github.com/yoojia/go-jsonx"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func cmdToJSON(cmd *dongk.Command) (bytes []byte, doorId, direct byte) {
	// 控制指令数据：
	// 0. 无记录
	// 1. 刷卡消息
	// 2. 门磁、按钮、设备启动、远程开门记录
	// 3. 报警记录
	// Reader一个按顺序读取字节数组的包装类
	reader := edgex.WrapByteReader(cmd.Data[:], binary.LittleEndian)
	json := jsonx.NewFatJSON()
	json.Field("sn", cmd.SerialNum)
	card := reader.GetBytesSize(4)
	json.Field("card", binary.LittleEndian.Uint32(card))
	json.Field("cardHex", hex.EncodeToString(card))
	reader.GetBytesSize(7) // 丢弃timestamp数据
	json.Field("index", reader.GetUint32())
	json.Field("type", reader.GetByte())
	json.Field("state", reader.GetByte())
	doorNo := reader.GetByte()
	ioDirect := reader.GetByte()
	json.Field("doorId", doorNo)
	json.Field("direct", dongk.DirectName(ioDirect))
	json.Field("reason", reader.GetByte())
	return json.Bytes(), doorNo, ioDirect
}
