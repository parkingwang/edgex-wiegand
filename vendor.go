package wiegand

import "fmt"

//
// Author: 陈哈哈 bitschen@163.com
//

const (
	VendorName       = "Wiegand"
	ConnectionDriver = "UDP"
)

func makeGroupId(serialNum uint32) string {
	return fmt.Sprintf("SNID[%d]", serialNum)
}

func makeMajorId(doorId int) string {
	return fmt.Sprintf("SNID[%d]", doorId)
}
