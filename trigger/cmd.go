package main

import (
	"fmt"
	"github.com/nextabc-lab/edgex-wiegand"
	"github.com/parkingwang/go-wg26"
	"github.com/yoojia/go-jsonx"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

// 19 20 0000
// cd6c4d0d
// 01180000
// 01
// 00
// 01
// 01
// fb 7c 83 00
// 2019 06 13 17 070306010100000000000000170703000000000000000000000019061300000000000000000000
func cmdToJSON(cmd *wiegand.Command) (frames []byte, cardNum string, doorId, direct, rType byte) {
	// 控制指令数据
	json := jsonx.NewFatJSON()
	json.Field("sn", cmd.SerialNum)
	// Reader一个按顺序读取字节数组的包装类
	reader := cmd.DataReader()
	json.Field("index", reader.NextUint32())
	rType = reader.NextByte()
	json.Field("type", rType)
	json.Field("typeName", wiegand.TypeName(rType))
	json.Field("state", reader.NextByte())
	doorId = reader.NextByte()
	direct = reader.NextByte()
	// 卡号字段是维根26码字面数值
	card := reader.NextUint32()
	id := wg26.ParseFromWg26Number(fmt.Sprintf("%d", card))
	json.Field("card", id.CardSN)
	json.Field("doorId", doorId)
	json.Field("direct", wiegand.DirectName(direct))
	return json.Bytes(), id.CardSN, doorId, direct, rType
}
