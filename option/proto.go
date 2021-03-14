// +build !kcp

package option

import (
	"net"
	"time"
)

const SupportedProtocols = "TCP UDP PROXY"

func (sd SocketDesc) IsProtoSupported() bool {
	switch sd.Proto {
	case "tcp", "udp", "proxy":
		return true
	default:
		return false
	}
}

func (sd SocketDesc) innerGetListener() (net.Listener, error) {
	switch {
	case sd.Proto[:3] == "tcp":
		return net.Listen(sd.Proto, sd.Addr)
	case sd.Proto == "proxy":
		return net.Listen("tcp", sd.Addr)
	}
	return nil, nil
}

func (sd SocketDesc) innerGetConn() (net.Conn, error) {
	switch {
	case sd.Proto[:3] == "tcp":
		return net.DialTimeout(sd.Proto, sd.Addr, time.Duration(TIMEOUT)*time.Millisecond)
	}
	return nil, nil
}
