package wiegand

import (
	"fmt"
	"github.com/nextabc-lab/edgex-go/extra"
)

//
// Author: 陈哈哈 bitschen@163.com
//

const (
	VendorName       = "Wiegand"
	ConnectionDriver = "UDP"
)

// AttrKey，用来标记[主板、门号]级别的EventId。
// 不要使用到MinorId，否则会出现：同一个门的READER、SWITCH不匹配的情况
func attrKeyRpcEventId(boardId, majorId string) string {
	return "ATTR/" + boardId + "/" + majorId + "/EVENT_ID`"
}

func makeBoardId(serialNum uint32) string {
	return fmt.Sprintf("SNID[%d]", serialNum)
}

func makeMajorId(doorId int) string {
	return fmt.Sprintf("DOOR[%d]", doorId)
}

// 解析状态
func stateName(s byte) string {
	// 有效性(0 表示不通过, 1 表示通过)
	if 1 == s {
		return "PASSED"
	} else {
		return "REJECT"
	}
}

// 解析进出方向类型名称
func directVal(b byte) byte {
	// 1: In, 2: Out
	if 1 == b {
		return extra.DirectIn
	} else {
		return extra.DirectOut
	}
}

func directName(b byte) string {
	// 1: In, 2: Out
	if 1 == b {
		return "IN"
	} else {
		return "OUT"
	}
}

// 解析记录类型名称
func typeName(b byte) string {
	switch b {
	case 0:
		return extra.TypeNop // 无记录

	case 1:
		return extra.TypeCard // 刷卡

	case 2:
		return extra.TypeOpen // 门磁，按钮，设备启动，远程开门

	case 3:
		return extra.TypeAlarm // 报警

	default:
		return extra.TypeNop
	}
}
