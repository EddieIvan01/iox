package option

import (
	"errors"
	"net"
	"time"

	"github.com/xtaci/smux"
)

type SocketDesc struct {
	Proto string
	Addr  string
	raw   string

	smuxSession *smux.Session

	Secret       bool
	Compress     bool
	Multiplexing bool
	IsListener   bool
}

func (sd SocketDesc) IsProtoReliable() bool {
	return sd.Proto[:3] != "udp"
}

func (sd SocketDesc) IsProxyProto() bool {
	return sd.Proto == "proxy" || sd.Proto == "kproxy"
}

func (sd SocketDesc) String() string {
	return sd.raw
}

func (sd *SocketDesc) GetListener() (net.Listener, error) {
	listener, err := sd.innerGetListener()
	if err != nil {
		return nil, err
	}

	if sd.Multiplexing {
		return sd.getSmuxServer(listener)
	}

	return listener, err
}

func (sd *SocketDesc) getSmuxServer(listener net.Listener) (net.Listener, error) {
	if sd.smuxSession == nil {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			sd.smuxSession, err = smux.Server(conn, &smux.Config{
				Version:           2,
				KeepAliveDisabled: false,
				KeepAliveInterval: SMUX_KEEPALIVE_INTERVAL * time.Second,
				KeepAliveTimeout:  SMUX_KEEPALIVE_TIMEOUT * time.Second,
				MaxFrameSize:      SMUX_FRAMESIZE,
				MaxReceiveBuffer:  SMUX_RECVBUFFER,
				MaxStreamBuffer:   SMUX_STREAMBUFFER,
			})
			if err != nil {
				continue
			}
			break
		}
	}

	return sd.smuxSession, nil
}

func (sd *SocketDesc) GetConn() (net.Conn, error) {
	if sd.smuxSession != nil {
		return sd.smuxSession.OpenStream()
	}

	conn, err := sd.innerGetConn()
	if err != nil {
		return nil, err
	}

	if sd.Multiplexing {
		return sd.getSmuxConn(conn)
	}

	return conn, err
}

func (sd *SocketDesc) getSmuxConn(conn net.Conn) (net.Conn, error) {
	var err error
	sd.smuxSession, err = smux.Client(conn, &smux.Config{
		Version:           2,
		KeepAliveDisabled: false,
		KeepAliveInterval: SMUX_KEEPALIVE_INTERVAL * time.Second,
		KeepAliveTimeout:  SMUX_KEEPALIVE_TIMEOUT * time.Second,
		MaxFrameSize:      SMUX_FRAMESIZE,
		MaxReceiveBuffer:  SMUX_RECVBUFFER,
		MaxStreamBuffer:   SMUX_STREAMBUFFER,
	})
	if err != nil {
		return nil, err
	}

	return sd.smuxSession.OpenStream()
}

func (sd SocketDesc) GetUDPConn() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr(sd.Proto, sd.Addr)
	if err != nil {
		return nil, err
	}

	if sd.IsListener {
		listener, err := net.ListenUDP(sd.Proto, addr)
		if err != nil {
			return nil, err
		}
		return listener, nil
	}

	conn, err := net.DialUDP(sd.Proto, nil, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil

}

func NewSocketDesc(s string) (*SocketDesc, error) {
	sd := &SocketDesc{
		raw: s,
	}
	s = stringToLower(s)

	if index := bindex(s, '@'); index >= 0 {
		for _, c := range s[:index] {
			switch c {
			case 'x':
				sd.Multiplexing = true
			case 'c':
				sd.Compress = true
			case 's':
				sd.Secret = true
			default:
				return nil, errors.New("Unknown option: " + string(c))
			}
		}

		s = s[index+1:]
	}

	index := bindex(s, ':')
	if index < 0 {
		return nil, errors.New("Invalid socket descriptor")
	}

	if s[index-2:index] == "-l" {
		sd.Proto = s[:index-2]
		sd.IsListener = true
	} else {
		sd.Proto = s[:index]
	}

	if !sd.IsProtoSupported() {
		panic("Unsupported protocol: " + sd.Proto)
	}

	sd.Addr = s[index+1:]
	if bindex(sd.Addr, ':') == -1 {
		sd.Addr = ":" + sd.Addr
	}

	return sd, nil
}

func stringToLower(s string) string {
	bs := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c ^= 0x20
		}
		bs[i] = c
	}
	return string(bs)
}

func bindex(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
