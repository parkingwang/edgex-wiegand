package wiegand

import (
	"encoding/binary"
	"errors"
	"github.com/yoojia/go-bytes"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
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

// 微耕Magic位
const (
	VendorMagic = 0x17
)

// 开关控制方式
const (
	SwitchDelayAlwaysOpen   = 0x01 // 常开
	SwitchDelayAlwaysClose  = 0x02 // 常闭
	SwitchDelayAlwaysOnline = 0x03 // 在线控制，默认方式
)

// 微耕门禁主板指令
type Command struct {
	Magic     byte     // 1
	FuncId    byte     // 1
	reversed  uint16   // 2
	SerialNum uint32   // 4
	Data      [32]byte // 32
	SeqId     uint32   // 4
	Extra     [20]byte // 20
}

func (wgc *Command) Bytes() []byte {
	// 微耕数据包使用小字节序
	br := bytes.NewWriter(ByteOrder)
	br.NextByte(wgc.Magic)
	br.NextByte(wgc.FuncId)
	br.NextUint16(wgc.reversed)
	br.NextUint32(wgc.SerialNum)
	br.NextBytes(wgc.Data[:])
	br.NextUint32(wgc.SeqId)
	br.NextBytes(wgc.Extra[:])
	return br.Bytes()
}

// Success 返回接收报文的成功标记位状态
func (wgc *Command) Success() bool {
	return 0x01 == wgc.Data[0]
}

func (wgc *Command) DataReader() *bytes.Reader {
	return bytes.WrapReader(wgc.Data[:], ByteOrder)
}

// 创建指令
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

// 创建指令
func NewCommand(funcId byte, serialId uint32, seqId uint32, data [32]byte) *Command {
	return NewCommand0(
		VendorMagic,
		funcId,
		0x00,
		serialId,
		seqId,
		data,
		[20]byte{})
}

// 解析数据指令。
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
