package main

import (
	"github.com/pkg/errors"
	"net"
	"time"
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

func tryWrite(conn *net.UDPConn, bs []byte, to time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(to)); nil != err {
		return err
	}
	if _, err := conn.Write(bs); nil != err {
		return err
	} else {
		return nil
	}
}

func tryRead(conn *net.UDPConn, buffer []byte, to time.Duration) (n int, err error) {
	if err := conn.SetReadDeadline(time.Now().Add(to)); nil != err {
		return 0, err
	}
	return conn.Read(buffer)
}
