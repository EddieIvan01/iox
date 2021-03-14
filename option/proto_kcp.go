// +build kcp

package option

import (
	"net"
	"time"

	"github.com/xtaci/kcp-go"
)

const SupportedProtocols = "TCP UDP KCP PROXY KPROXY"

func (sd SocketDesc) IsProtoSupported() bool {
	switch sd.Proto {
	case "tcp", "udp", "kcp", "proxy", "kproxy":
		return true
	default:
		return false
	}
}

func (sd SocketDesc) innerGetListener() (net.Listener, error) {
	switch {
	case sd.Proto[:3] == "tcp":
		return net.Listen(sd.Proto, sd.Addr)
	case sd.Proto == "kcp":
		return kcp.Listen(sd.Addr)
	case sd.Proto == "proxy":
		return net.Listen("tcp", sd.Addr)
	case sd.Proto == "kproxy":
		return kcp.Listen(sd.Addr)
	}
	return nil, nil
}

func (sd SocketDesc) innerGetConn() (net.Conn, error) {
	switch {
	case sd.Proto[:3] == "tcp":
		return net.DialTimeout(sd.Proto, sd.Addr, time.Duration(TIMEOUT)*time.Millisecond)
	case sd.Proto == "kcp":
		return kcp.Dial(sd.Addr)
	}
	return nil, nil
}
