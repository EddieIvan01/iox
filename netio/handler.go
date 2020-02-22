package netio

import (
	"iox/crypto"
	"iox/option"
	"net"
)

type Ctx interface {
	DecryptRead(b []byte) (int, error)
	EncryptWrite(b []byte) (int, error)
	IsEncrypted() bool

	net.Conn
}

type TCPCtx struct {
	net.Conn
	encrypted bool

	// Ensure synchronous
	encCipher *crypto.Cipher
	decCipher *crypto.Cipher
}

func NewTCPCtx(conn net.Conn, encrypted bool) (*TCPCtx, error) {
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetLinger(0)
	}

	ctx := &TCPCtx{
		Conn:      conn,
		encrypted: encrypted,
	}

	encCipher, decCipher, err := crypto.NewCipherPair(option.KEY)
	if err != nil {
		return nil, err
	}

	if encrypted {
		ctx.encCipher = encCipher
		ctx.decCipher = decCipher
	}

	return ctx, nil
}

func (c *TCPCtx) DecryptRead(b []byte) (int, error) {
	n, err := c.Read(b)
	if err != nil {
		return n, err
	}

	c.decCipher.StreamXOR(b[:n], b[:n])

	return n, err
}

func (c *TCPCtx) EncryptWrite(b []byte) (int, error) {
	c.encCipher.StreamXOR(b, b)
	return c.Write(b)
}

func (c TCPCtx) IsEncrypted() bool {
	return c.encrypted
}

type UDPCtx struct {
	net.Conn
	encrypted bool

	encCipher *crypto.Cipher
	decCipher *crypto.Cipher
}
