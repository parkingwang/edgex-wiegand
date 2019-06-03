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

func cmdToJSON(cmd *dongk.Command) (frames []byte, card string, doorId, direct byte) {
	// 控制指令数据：
	// 0. 无记录
	// 1. 刷卡消息
	// 2. 门磁、按钮、设备启动、远程开门记录
	// 3. 报警记录
	json := jsonx.NewFatJSON()
	json.Field("sn", cmd.SerialNum)
	// Reader一个按顺序读取字节数组的包装类
	reader := bytes.WrapReader(cmd.Data[:], dongk.ByteOrder)
	json.Field("index", reader.NextUint32())
	json.Field("type", reader.NextByte())
	json.Field("state", reader.NextByte())
	doorId = reader.NextByte()
	direct = reader.NextByte()
	card = hex.EncodeToString(reader.NextBytes(4))
	json.Field("card", card)
	reader.NextBytes(7) // 丢弃timestamp数据
	json.Field("doorId", doorId)
	json.Field("direct", dongk.DirectName(direct))
	return json.Bytes(), card, doorId, direct
}
