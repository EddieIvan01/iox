package operate

import (
	"errors"
	"io"
	"iox/logger"
	"iox/option"
	"net"
	"time"
)

const (
	CTL_HANDSHAKE = iota
	CTL_CONNECT_ME
	CTL_CLEANUP
	CTL_HEARTBEAT

	MAX_CONNECTION   = 0x400
	CLIENT_HANDSHAKE = 0xC0
	SERVER_HANDSHAKE = 0xE0
)

type Protocol struct {
	CMD byte
	N   byte

	// ACK uint16
}

var END = []byte{0xEE, 0xFF}

func marshal(p Protocol) []byte {
	buf := make([]byte, 4)
	buf[0] = p.CMD
	buf[1] = p.N

	buf[2], buf[3] = END[0], END[1]
	return buf
}

func unmarshal(b []byte) (*Protocol, error) {
	if len(b) < 2 {
		return nil, errors.New("Protocol data is too short")
	}

	p := &Protocol{
		CMD: b[0],
		N:   b[1],
	}

	return p, nil
}

func bytesEq(a, b []byte) bool {
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func readUntilEnd(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 1)
	output := make([]byte, 0, 4)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if n != 1 {
			return nil, errors.New("Transmission error")
		}

		output = append(output, buf[0])

		if len(output) >= 2 && bytesEq(END, output[len(output)-2:len(output)]) {
			break
		}
	}

	return output[:2], nil
}

func serverHandshake(listener net.Listener) net.Conn {
	var masterConn net.Conn
	var err error
	for {
		masterConn, err = listener.Accept()
		if err != nil {
			continue
		}

		pb, err := readUntilEnd(masterConn)
		if err != nil {
			continue
		}

		p, err := unmarshal(pb)
		if err != nil {
			continue
		}

		if p.CMD == CTL_HANDSHAKE && p.N == CLIENT_HANDSHAKE {
			logger.Success("Remote socks5 handshake ok")
			masterConn.Write(marshal(Protocol{
				CMD: CTL_HANDSHAKE,
				N:   SERVER_HANDSHAKE,
			}))
			break
		}
	}

	return masterConn
}

func clientHandshake(remote string) (net.Conn, error) {
	masterConn, err := net.DialTimeout(
		"tcp", remote,
		time.Millisecond*time.Duration(option.TIMEOUT),
	)
	if err != nil {
		return nil, err
	}

	masterConn.Write(marshal(Protocol{
		CMD: CTL_HANDSHAKE,
		N:   CLIENT_HANDSHAKE,
	}))

	pb, err := readUntilEnd(masterConn)
	if err != nil {
		return nil, errors.New("Connect to remote forward server error")
	}

	p, err := unmarshal(pb)
	if err != nil {
		return nil, errors.New("Connect to remote forward server error")
	}
	if p.CMD == CTL_HANDSHAKE && p.N == SERVER_HANDSHAKE {
		logger.Success("Connect to remote forward server ok")
	} else {
		return nil, errors.New("Connect to remote forward server error")
	}

	return masterConn, nil
}
