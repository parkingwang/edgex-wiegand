package main

import (
	"encoding/hex"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/yoojia/go-bytes"
	"github.com/yoojia/go-jsonx"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func cmdToJSON(cmd *dongk.Command) (frames []byte, doorId, direct byte) {
	// 控制指令数据：
	// 0. 无记录
	// 1. 刷卡消息
	// 2. 门磁、按钮、设备启动、远程开门记录
	// 3. 报警记录
	// Reader一个按顺序读取字节数组的包装类
	reader := bytes.WrapReader(cmd.Data[:], dongk.ByteOrder)
	json := jsonx.NewFatJSON()
	json.Field("sn", cmd.SerialNum)
	card := reader.NextBytes(4)
	json.Field("card", dongk.ByteOrder.Uint32(card))
	json.Field("cardHex", hex.EncodeToString(card))
	reader.NextBytes(7) // 丢弃timestamp数据
	json.Field("index", reader.NextUint32())
	json.Field("type", reader.NextByte())
	json.Field("state", reader.NextByte())
	doorNo := reader.NextByte()
	ioDirect := reader.NextByte()
	json.Field("doorId", doorNo)
	json.Field("direct", dongk.DirectName(ioDirect))
	json.Field("reason", reader.NextByte())
	return json.Bytes(), doorNo, ioDirect
}
