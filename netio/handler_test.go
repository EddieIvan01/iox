package netio

import (
	"net"
	"testing"
)

func bytesEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestTCPCtx(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Error(err.Error())
	}
	defer listener.Close()

	buf := make([]byte, 1024)
	signal := make(chan struct{}, 1)
	msg := "testing message."

	go func() {
		server, err := listener.Accept()
		if err != nil {
			t.Error(err.Error())
		}
		defer server.Close()

		serverCtx, _ := NewTCPCtx(server, true)
		serverCtx.DecryptRead(buf)
		signal <- struct{}{}
	}()

	client, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Error(err.Error())
	}
	defer client.Close()

	clientCtx, err := NewTCPCtx(client, true)
	if err != nil {
		t.Error(err.Error())
	}
	clientCtx.EncryptWrite([]byte(msg))

	<-signal
	if !bytesEq([]byte(msg), buf[:len(msg)]) {
		t.Error("TCPCtx error")
	}
}

func TestUDPConn(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", ":9999")
	l, _ := net.ListenUDP("udp", addr)
	lCtx, _ := NewUDPCtx(l, true, false)

	signal := make(chan struct{}, 0)

	go func() {
		addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
		c, _ := net.DialUDP("udp", nil, addr)
		cCtx, _ := NewUDPCtx(c, true, true)

		cCtx.EncryptWrite([]byte("testing message."))

		signal <- struct{}{}
	}()

	<-signal

	buf := make([]byte, 32)
	n, err := lCtx.DecryptRead(buf)
	if err != nil {
		t.Error(err.Error())
	}

	if string(buf[:n]) != "testing message." {
		t.Log(buf[:n])
		t.Error("UDPCtx Error")
	}
}
