package main

import (
	"github.com/bitschen/go-jsonx"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/parkingwang/go-wg26"
)

//
// Author: 陈哈哈 bitschen@163.com
//

func cmdToJSON(cmd *dongk.Command) (frames []byte, cardNum string, doorId, direct, rType byte) {
	// 控制指令数据
	json := jsonx.NewFatJSON()
	json.Field("sn", cmd.SerialNum)
	// Reader一个按顺序读取字节数组的包装类
	reader := cmd.DataReader()
	json.Field("index", reader.NextUint32())
	rType = reader.NextByte()
	json.Field("type", rType)
	json.Field("typeName", dongk.TypeName(rType))
	json.Field("state", reader.NextByte())
	doorId = reader.NextByte()
	direct = reader.NextByte()
	id := wg26.ParseFromUint32(reader.NextUint32())
	json.Field("card", id.Number)
	reader.NextBytes(7) // 丢弃timestamp数据
	json.Field("doorId", doorId)
	json.Field("direct", dongk.DirectName(direct))
	return json.Bytes(), id.Number, doorId, direct, rType
}
