package wiegand

import (
	"errors"
	"fmt"
	"github.com/albenik/bcd"
	"github.com/parkingwang/go-wg26"
	"github.com/yoojia/go-at"
	"github.com/yoojia/go-bytes"
	"strconv"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

func AtCommands(registry *at.AtRegister, broadSN uint32) {
	// AT+OPEN=SWITCH_ID
	registry.AddX("OPEN", 1, func(args ...string) ([]byte, error) {
		switchId, err := parseInt(args[0])
		if nil != err {
			return nil, errors.New("INVALID_SWITCH_ID:" + args[0])
		}
		return NewCommand(FunIdRemoteOpen,
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
		return NewCommand(FunIdSwitchDelay,
				broadSN,
				0,
				[32]byte{
					byte(switchId),          // 门号
					SwitchDelayAlwaysOnline, // 控制方式
					byte(seconds),           //开门延时：秒
				}).Bytes(),
			nil
	})
	// AT+ADD=CARD(ID),START_DATE(YYYYMMdd),END_DATE(YYYYMMdd),DOOR1,DOOR2,DOOR3,DOOR4
	addHandler := func(args ...string) ([]byte, error) {
		cardSN := args[0]
		if !wg26.IsCardSN(cardSN) {
			return nil, errors.New("INVALID_CARD_SN[10digits]")
		}
		// 东控卡号使用的是 WG26SN 对应的字面数值
		w := bytes.NewWriter(ByteOrder)
		w.NextUint32(wg26.ParseFromCardNumber(cardSN).ValueOfWg26SN())
		w.NextBytes(getDateOrDefault(args, 1, 20190101))
		w.NextBytes(getDateOrDefault(args, 2, 20291231)) // 20290101
		w.NextByte(byte(getIntOrDefault(args, 3, 1)))
		w.NextByte(byte(getIntOrDefault(args, 4, 1)))
		w.NextByte(byte(getIntOrDefault(args, 5, 1)))
		w.NextByte(byte(getIntOrDefault(args, 6, 1)))
		data := [32]byte{}
		copy(data[:], w.Bytes())
		return NewCommand(FunIdCardAdd, broadSN, 0, data).Bytes(),
			nil
	}
	registry.AddX("ADD", 1, addHandler)
	registry.Add("ADD0", addHandler)

	// AT+DELETE=CARD(ID)
	registry.AddX("DELETE", 1, func(args ...string) ([]byte, error) {
		cardSN := args[0]
		if !wg26.IsCardSN(cardSN) {
			return nil, errors.New("INVALID_CARD_SN[10digits]")
		}
		// 卡号使用的是 WG26SN 对应的字面数值
		w := bytes.NewWriter(ByteOrder)
		w.NextUint32(wg26.ParseFromCardNumber(cardSN).ValueOfWg26SN())
		data := [32]byte{}
		copy(data[:], w.Bytes())
		return NewCommand(FunIdCardDel, broadSN, 0, data).Bytes(),
			nil
	})

	// AT+CLEAR
	registry.AddX("CLEAR", 0, func(args ...string) ([]byte, error) {
		return NewCommand(FunIdCardClear, broadSN, 0, [32]byte{0x55, 0xAA, 0xAA, 0x55}).Bytes(),
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

func parseInt(val string) (int64, error) {
	return strconv.ParseInt(val, 10, 64)
}
