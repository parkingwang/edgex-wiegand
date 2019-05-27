package dongk

import (
	"encoding/binary"
	"errors"
	"github.com/yoojia/go-bytes"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 传递指令时使用的字节顺序
var ByteOrder = binary.LittleEndian

const (
	// 主板状态
	FunIdBoardState = 0x20
	// 远程开门
	FunIdRemoteOpen = 0x40
	// 设置开关延时
	FunIdSwitchDelay = 0x80
	// 添加卡
	FunIdCardAdd = 0x50
	// 删除卡
	FunIdCardDel = 0x52
	// 删除所有卡
	FunIdCardClear = 0x54
)

// 东控Magic位
const (
	Magic = 0x19
)

// 开关控制方式
const (
	SwitchDelayAlwaysOpen   = 0x01 // 常开
	SwitchDelayAlwaysClose  = 0x02 // 常闭
	SwitchDelayAlwaysOnline = 0x03 // 在线控制，默认方式
)

const (
	DirectIn  = 1
	DirectOut = 2
)

// 东控门禁主板指令
type Command struct {
	Magic     byte     // 1
	FuncId    byte     // 1
	reversed  uint16   // 2
	SerialNum uint32   // 4
	Data      [32]byte // 32
	SeqId     uint32   // 4
	Extra     [20]byte // 20
}

func (dk *Command) Bytes() []byte {
	// 东控数据包使用小字节序
	br := bytes.NewWriter(ByteOrder)
	br.NextByte(dk.Magic)
	br.NextByte(dk.FuncId)
	br.NextUint16(dk.reversed)
	br.NextUint32(dk.SerialNum)
	br.NextBytes(dk.Data[:])
	br.NextUint32(dk.SeqId)
	br.NextBytes(dk.Extra[:])
	return br.Bytes()
}

// Success 返回接收报文的成功标记位状态
func (dk *Command) Success() bool {
	return 0x01 == dk.Data[0]
}

// 创建DK指令
func NewCommand0(magic, funcId byte, nop uint16, serial uint32, seqId uint32, data [32]byte, extra [20]byte) *Command {
	return &Command{
		Magic:     magic,
		FuncId:    funcId,
		reversed:  nop,
		SerialNum: serial,
		Data:      data,
		SeqId:     seqId,
		Extra:     extra,
	}
}

// 创建DK指令
func NewCommand(funcId byte, serialId uint32, seqId uint32, data [32]byte) *Command {
	return NewCommand0(
		Magic,
		funcId,
		0x00,
		serialId,
		seqId,
		data,
		[20]byte{})
}

// 解析DK数据指令。
func ParseCommand(frame []byte) (*Command, error) {
	if len(frame) < 9 {
		return nil, errors.New("invalid bytes len")
	}
	br := bytes.WrapReader(frame, ByteOrder)
	magic := br.NextByte()
	funId := br.NextByte()
	reserved := br.NextUint16()
	serialNum := br.NextUint32()
	data := [32]byte{}
	copy(data[:], br.NextBytes(32))
	seqId := br.NextUint32()
	extra := [20]byte{}
	copy(extra[:], br.NextBytes(20))
	return NewCommand0(
		magic,
		funId,
		reserved,
		serialNum,
		seqId,
		data,
		extra,
	), nil
}

func DirectName(b byte) string {
	if 1 == b {
		return "IN"
	} else {
		return "OUT"
	}
}
