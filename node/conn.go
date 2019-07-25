package main

import (
	"github.com/pkg/errors"
	"net"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

func makeUdpConn(remoteAddr string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if nil != err {
		return nil, errors.WithMessage(err, "Resolve udp address failed")
	}
	return net.DialUDP("udp", nil, addr)
}
