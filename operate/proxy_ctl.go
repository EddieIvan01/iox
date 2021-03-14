package operate

import (
	"errors"
	"io"
	"iox/netio"
	"iox/option"
	"net"
)

const (
	CTL_HANDSHAKE = iota
	CTL_CONNECT_ME
	CTL_CLEANUP

	MAX_CONNECTION   = 0x800
	CLIENT_HANDSHAKE = 0xc0
	SERVER_HANDSHAKE = 0xe0
)

type Protocol struct {
	CMD byte
	N   byte

	// ACK uint16
}

var PROTO_END = []byte{0xee, 0xff}

func marshal(p Protocol) []byte {
	buf := make([]byte, 4)
	buf[0] = p.CMD
	buf[1] = p.N

	buf[2], buf[3] = PROTO_END[0], PROTO_END[1]
	return buf
}

func unmarshal(b []byte) Protocol {
	return Protocol{
		CMD: b[0],
		N:   b[1],
	}
}

func bytesEq(a, b []byte) bool {
	return string(a) == string(b)
}

func readUntilEnd(reader io.Reader) ([]byte, error) {
	buf := make([]byte, 1)
	output := make([]byte, 0, 4)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			return nil, err
		}

		if n != 1 || len(output) > 4 {
			return nil, errors.New("Transmission error")
		}

		output = append(output, buf[0])

		if len(output) == 4 && bytesEq(PROTO_END, output[len(output)-2:]) {
			break
		}
	}
	return output[:2], nil
}

func serverHandshake(listener net.Listener, encrypted bool, compress bool) (*netio.TCPCtx, error) {
	var ctlStream net.Conn
	var ctlStreamCtx *netio.TCPCtx
	var err error

	for {
		ctlStream, err = listener.Accept()
		if err != nil {
			return nil, err
		}
		ctlStreamCtx, err = netio.NewTCPCtx(ctlStream, encrypted, compress)
		if err != nil {
			return ctlStreamCtx, err
		}

		pb, err := readUntilEnd(ctlStreamCtx)
		if err != nil {
			continue
		}

		p := unmarshal(pb)
		if p.CMD == CTL_HANDSHAKE && p.N == CLIENT_HANDSHAKE {
			_, err = ctlStreamCtx.Write(marshal(Protocol{
				CMD: CTL_HANDSHAKE,
				N:   SERVER_HANDSHAKE,
			}))
			if err != nil {
				return ctlStreamCtx, err
			}
			break
		}
	}

	return ctlStreamCtx, nil
}

func clientHandshake(remoteDesc *option.SocketDesc) (*netio.TCPCtx, error) {
	ctlStream, err := remoteDesc.GetConn()
	if err != nil {
		return nil, err
	}

	ctlStreamCtx, err := netio.NewTCPCtx(ctlStream, remoteDesc.Secret, remoteDesc.Compress)
	if err != nil {
		return ctlStreamCtx, err
	}

	_, err = ctlStreamCtx.Write(marshal(Protocol{
		CMD: CTL_HANDSHAKE,
		N:   CLIENT_HANDSHAKE,
	}))
	if err != nil {
		return ctlStreamCtx, err
	}

	pb, err := readUntilEnd(ctlStreamCtx)
	if err != nil {
		return ctlStreamCtx, errors.New("Connect to remote forward server error")
	}

	p := unmarshal(pb)
	if !(p.CMD == CTL_HANDSHAKE && p.N == SERVER_HANDSHAKE) {
		return ctlStreamCtx, errors.New("Connect to remote forward server error")
	}

	return ctlStreamCtx, nil
}
