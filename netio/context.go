package netio

import (
	"io"
	"iox/crypto"
	"iox/option"
	"net"

	"github.com/klauspost/compress/s2"
)

type Ctx interface {
	io.ReadWriteCloser
}

var _ Ctx = &TCPCtx{}
var _ Ctx = &UDPCtx{}

type TCPCtx struct {
	net.Conn

	reader io.Reader
	writer io.Writer

	writerFlush func() error

	encrypted bool
	compress  bool
}

// The returned *TCPCtx will never be nil
func NewTCPCtx(conn net.Conn, encrypted bool, compress bool) (*TCPCtx, error) {
	encrypted = encrypted && !option.FORWARD_WITHOUT_DEC
	compress = compress && !option.FORWARD_WITHOUT_COMPRESS

	ctx := &TCPCtx{
		Conn: conn,

		reader: conn,
		writer: conn,

		encrypted: encrypted,
	}

	if encrypted {
		encIV, err := crypto.RandomNonce()
		if err != nil {
			return ctx, err
		}
		_, err = conn.Write(encIV)
		if err != nil {
			return ctx, err
		}

		decIV := make([]byte, 0x18)
		_, err = io.ReadFull(conn, decIV)
		if err != nil {
			return ctx, err
		}

		ctx.reader, err = crypto.NewReader(ctx.reader, decIV)
		if err != nil {
			return ctx, err
		}
		ctx.writer, err = crypto.NewWriter(ctx.writer, encIV)
		if err != nil {
			return ctx, err
		}
	}

	if compress {
		ctx.reader = GetCompressReader(ctx.reader)

		writer := GetCompressWriter(ctx.writer)
		ctx.writerFlush = writer.Flush
		ctx.writer = writer
	}

	ctx.compress = compress
	return ctx, nil
}

func (ctx *TCPCtx) Read(b []byte) (int, error) {
	return ctx.reader.Read(b)
}

func (ctx *TCPCtx) Write(b []byte) (int, error) {
	n, err := ctx.writer.Write(b)
	if ctx.writerFlush != nil {
		err = ctx.writerFlush()
		if err != nil {
			return 0, err
		}
	}
	return n, err
}

func (ctx *TCPCtx) Close() error {
	if ctx.compress {
		PutCompressReader(ctx.reader.(*s2.Reader))
		PutCompressWriter(ctx.writer.(*s2.Writer))
	}

	if ctx.Conn != nil {
		return ctx.Conn.Close()
	}
	return nil
}

type UDPCtx struct {
	*net.UDPConn
	remoteAddr *net.UDPAddr

	encrypted bool
	connected bool

	// sync.Mutex
}

func NewUDPCtx(conn *net.UDPConn, encrypted bool, connected bool) (*UDPCtx, error) {
	encrypted = encrypted && !option.FORWARD_WITHOUT_DEC

	ctx := &UDPCtx{
		UDPConn:   conn,
		encrypted: encrypted,
		connected: connected,
	}

	return ctx, nil
}

// Encryption for packet is different from stream
func (ctx *UDPCtx) Read(b []byte) (int, error) {
	var n int
	var err error

	if !ctx.connected {
		var remoteAddr *net.UDPAddr
		n, remoteAddr, err = ctx.UDPConn.ReadFromUDP(b)
		if err != nil {
			return n, err
		}
		ctx.remoteAddr = remoteAddr

	} else {
		n, err = ctx.UDPConn.Read(b)
		if err != nil {
			return n, err
		}
	}

	if ctx.encrypted {
		if len(b) < 0x18 {
			// no nonce, skip
			return 0, nil
		}
		nonce := b[n-0x18 : n]
		b = b[:n-0x18]

		cipher, err := crypto.NewCipher(nonce)
		if err != nil {
			return 0, err
		}

		n -= 0x18
		cipher.StreamXOR(b[:n], b[:n])
	}

	return n, err
}

func (ctx *UDPCtx) Write(b []byte) (int, error) {
	if ctx.encrypted {
		iv, err := crypto.RandomNonce()
		cipher, err := crypto.NewCipher(iv)
		if err != nil {
			return 0, err
		}

		cipher.StreamXOR(b, b)
		b = append(b, iv...)
	}

	if !ctx.connected {
		return ctx.UDPConn.WriteTo(b, ctx.remoteAddr)
	}
	return ctx.UDPConn.Write(b)
}

/*
func (c UDPCtx) IsRemoteAddrRegistered() bool {
	return c.remoteAddr != nil
}
*/
