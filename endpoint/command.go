package main

import (
	"errors"
	"fmt"
	"github.com/albenik/bcd"
	"github.com/nextabc-lab/edgex-dongkong"
	"github.com/yoojia/go-at"
	"github.com/yoojia/go-bytes"
	"strconv"
)

//
// Author: 陈哈哈 bitschen@163.com
//

func atCommands(registry *at.AtRegister, broadSN uint32) {
	// AT+OPEN=SWITCH_ID
	registry.AddX("OPEN", 1, func(args ...string) ([]byte, error) {
		switchId, err := parseInt(args[0])
		if nil != err {
			return nil, errors.New("INVALID_SWITCH_ID:" + args[0])
		}
		return dongk.NewCommand(dongk.FunIdRemoteOpen,
				broadSN,
				0,
				[32]byte{byte(switchId)}).Bytes(),
			nil
	})
	// AT+DELAY=SWITCH_ID,DELAY_SEC
	registry.AddX("DELAY", 2, func(args ...string) ([]byte, error) {
		switchId, err := parseInt(args[0])
		if nil != err {
			return nil, errors.New("INVALID_SWITCH_ID:" + args[0])
		}
		seconds, err := parseInt(args[1])
		if nil != err {
			return nil, errors.New("INVALID_DELAY_SEC:" + args[1])
		}
		if 0 > seconds || seconds > 60 {
			return nil, errors.New(fmt.Sprintf("INVALID_DELAY_SEC: %d", seconds))
		}
		return dongk.NewCommand(dongk.FunIdSwitchDelay,
				broadSN,
				0,
				[32]byte{
					byte(switchId),                // 门号
					dongk.SwitchDelayAlwaysOnline, // 控制方式
					byte(seconds),                 //开门延时：秒
				}).Bytes(),
			nil
	})
	// AT+ADD=CARD(uint32),START_DATE(YYYYMMdd),END_DATE(YYYYMMdd),DOOR1,DOOR2,DOOR3,DOOR4
	addHandler := func(args ...string) ([]byte, error) {
		card, err := getCardNumber(args[0])
		if nil != err {
			return nil, err
		}
		w := bytes.NewWriter(dongk.ByteOrder)
		w.NextUint32(card)
		w.NextBytes(getDateOrDefault(args, 1, 20190101))
		w.NextBytes(getDateOrDefault(args, 2, 20291231)) // 20290101
		w.NextByte(byte(getIntOrDefault(args, 3, 1)))
		w.NextByte(byte(getIntOrDefault(args, 4, 1)))
		w.NextByte(byte(getIntOrDefault(args, 5, 1)))
		w.NextByte(byte(getIntOrDefault(args, 6, 1)))
		data := [32]byte{}
		copy(data[:], w.Bytes())
		return dongk.NewCommand(dongk.FunIdCardAdd, broadSN, 0, data).Bytes(),
			nil
	}
	registry.AddX("ADD", 1, addHandler)
	registry.Add("ADD0", addHandler)

	// AT+DELETE=CARD(uint32)
	registry.AddX("DELETE", 1, func(args ...string) ([]byte, error) {
		card, err := getCardNumber(args[0])
		if nil != err {
			return nil, err
		}
		w := bytes.NewWriter(dongk.ByteOrder)
		w.NextUint32(card)
		data := [32]byte{}
		copy(data[:], w.Bytes())
		return dongk.NewCommand(dongk.FunIdCardDel, broadSN, 0, data).Bytes(),
			nil
	})

	// AT+CLEAR
	registry.AddX("CLEAR", 0, func(args ...string) ([]byte, error) {
		return dongk.NewCommand(dongk.FunIdCardClear, broadSN, 0, [32]byte{0x55, 0xAA, 0xAA, 0x55}).Bytes(),
			nil
	})
}

func getDateOrDefault(args []string, idx int, def uint32) []byte {
	uintDate := def
	if idx <= len(args)-1 {
		strDate := args[idx]
		if len("20190101") == len(strDate) {
			if val, err := parseInt(strDate); nil == err {
				uintDate = uint32(val)
			}
		}
	}
	return bcd.FromUint32(uintDate)
}

func getIntOrDefault(args []string, idx int, def uint32) uint32 {
	if idx > len(args)-1 {
		return def
	} else {
		if v, e := parseInt(args[idx]); nil != e {
			return def
		} else {
			return uint32(v)
		}
	}
}

func getCardNumber(val string) (uint32, error) {
	intCardNum, err := parseInt(val)
	if nil != err {
		return 0, errors.New("INVALID_CARD: " + val)
	} else {
		return uint32(intCardNum), nil
	}
}

func parseInt(val string) (int64, error) {
	return strconv.ParseInt(val, 10, 64)
}
