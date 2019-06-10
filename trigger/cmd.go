package main

import (
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/yoojia/go-bytes"
	"github.com/yoojia/go-jsonx"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func cmdToJSON(cmd *dongk.Command) (frames []byte, iCard uint32, doorId, direct, rType byte) {
	// 控制指令数据
	json := jsonx.NewFatJSON()
	json.Field("sn", cmd.SerialNum)
	// Reader一个按顺序读取字节数组的包装类
	reader := bytes.WrapReader(cmd.Data[:], dongk.ByteOrder)
	json.Field("index", reader.NextUint32())
	rType = reader.NextByte()
	json.Field("type", rType)
	json.Field("typeName", dongk.TypeName(rType))
	json.Field("state", reader.NextByte())
	doorId = reader.NextByte()
	direct = reader.NextByte()
	iCard = reader.NextUint32()
	json.Field("card", iCard)
	reader.NextBytes(7) // 丢弃timestamp数据
	json.Field("doorId", doorId)
	json.Field("direct", dongk.DirectName(direct))
	return json.Bytes(), iCard, doorId, direct, rType
}
