package netio

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

type _buffer struct {
	bytes.Buffer
}

func (b _buffer) IsEncrypted() bool                  { return false }
func (b _buffer) EncryptWrite([]byte) (int, error)   { return 0, nil }
func (b _buffer) DecryptRead([]byte) (int, error)    { return 0, nil }
func (b _buffer) Close() error                       { return nil }
func (b _buffer) LocalAddr() net.Addr                { return nil }
func (b _buffer) RemoteAddr() net.Addr               { return nil }
func (b _buffer) SetDeadline(t time.Time) error      { return nil }
func (b _buffer) SetReadDeadline(t time.Time) error  { return nil }
func (b _buffer) SetWriteDeadline(t time.Time) error { return nil }

func TestCipherCopy(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Error(err.Error())
	}
	defer listener.Close()

	buf := &_buffer{}

	signal := make(chan struct{}, 1)
	go func() {
		localConn, err := listener.Accept()
		if err != nil {
			t.Error(err.Error())
		}

		localConnCtx, err := NewTCPCtx(localConn, true)
		if err != nil {
			t.Error(err.Error())
		}

		CipherCopy(buf, localConnCtx)
		signal <- struct{}{}
	}()

	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Error(err.Error())
	}

	connCtx, err := NewTCPCtx(conn, true)
	if err != nil {
		t.Error(err.Error())
	}

	msg := "testing message."
	connCtx.EncryptWrite([]byte(msg))
	conn.Close()

	<-signal
	if buf.String() != msg {
		fmt.Println(buf.String())
		t.Error("CipherCopy error")
	}
}

func TestPipeForward(t *testing.T) {
	listenerA, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Error(err.Error())
	}
	defer listenerA.Close()

	listenerB, err := net.Listen("tcp", "127.0.0.1:8888")
	if err != nil {
		t.Error(err.Error())
	}
	defer listenerB.Close()

	var connA, connB net.Conn
	signal := make(chan struct{}, 1)

	msgA := "FROM A"
	msgB := "FROM B"

	bufA := make([]byte, 1024)
	bufB := make([]byte, 1024)

	go func() {
		localA, err := net.DialTimeout("tcp", "127.0.0.1:9999", time.Second*3)
		if err != nil {
			t.Error(err.Error())
		}
		defer localA.Close()

		localCtxA, err := NewTCPCtx(localA, true)
		if err != nil {
			t.Error(err.Error())
		}

		localCtxA.EncryptWrite([]byte(msgA))
		localCtxA.DecryptRead(bufA)
	}()

	go func() {
		localB, err := net.DialTimeout("tcp", "127.0.0.1:8888", time.Second*3)
		if err != nil {
			t.Error(err.Error())
		}
		defer localB.Close()

		localCtxB, err := NewTCPCtx(localB, true)
		if err != nil {
			t.Error(err.Error())
		}

		localCtxB.EncryptWrite([]byte(msgB))
		localCtxB.DecryptRead(bufB)
	}()

	go func() {
		var err error
		connA, err = listenerA.Accept()
		if err != nil {
			t.Error(err.Error())
		}
		signal <- struct{}{}
	}()

	go func() {
		var err error
		connB, err = listenerB.Accept()
		if err != nil {
			t.Error(err.Error())
		}
		signal <- struct{}{}
	}()

	<-signal
	<-signal

	connCtxA, err := NewTCPCtx(connA, true)
	if err != nil {
		t.Error(err.Error())
	}

	connCtxB, err := NewTCPCtx(connB, true)
	if err != nil {
		t.Error(err.Error())
	}

	PipeForward(connCtxA, connCtxB)

	if string(bufA[:len(msgB)]) != msgB || string(bufB[:len(msgA)]) != msgA {
		t.Error("PipeForward error")
	}
}
