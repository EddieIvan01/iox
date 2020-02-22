package operate

import (
	"iox/netio"
	"net"
	"testing"
	"time"
)

// run forever
func testLocal2Local(t *testing.T) {
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

		localCtxA, err := netio.NewTCPCtx(localA, true)
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

		localCtxB, err := netio.NewTCPCtx(localB, true)
		if err != nil {
			t.Error(err.Error())
		}

		localCtxB.EncryptWrite([]byte(msgB))
		localCtxB.DecryptRead(bufB)
	}()

	Local2Local(":9999", ":8888", true, true)

	if string(bufA[:len(msgB)]) != msgB || string(bufB[:len(msgA)]) != msgA {
		t.Error("Local2Local error")
	}
}

func TestRemote2Remote(t *testing.T) {
	msgA := "FROM A"
	msgB := "FROM B"

	bufA := make([]byte, 1024)
	bufB := make([]byte, 1024)

	go func() {
		listenerA, err := net.Listen("tcp", ":9999")
		if err != nil {
			t.Error(err.Error())
		}
		defer listenerA.Close()

		connA, err := listenerA.Accept()
		if err != nil {
			t.Error(err.Error())
		}
		defer connA.Close()

		connCtxA, err := netio.NewTCPCtx(connA, true)
		if err != nil {
			t.Error(err.Error())
		}

		connCtxA.EncryptWrite([]byte(msgA))
		connCtxA.DecryptRead(bufA)
	}()

	go func() {
		listenerB, err := net.Listen("tcp", ":8888")
		if err != nil {
			t.Error(err.Error())
		}
		defer listenerB.Close()

		connB, err := listenerB.Accept()
		if err != nil {
			t.Error(err.Error())
		}
		defer connB.Close()

		connCtxB, err := netio.NewTCPCtx(connB, true)
		if err != nil {
			t.Error(err.Error())
		}

		connCtxB.EncryptWrite([]byte(msgB))
		connCtxB.DecryptRead(bufB)
	}()

	Remote2Remote("127.0.0.1:9999", "127.0.0.1:8888", true, true)
	if string(bufA[:len(msgB)]) != msgB || string(bufB[:len(msgA)]) != msgA {
		t.Error("Remote2Remote error")
	}
}

// run forever
func testLocal2Remote(t *testing.T) {
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

		localCtxA, err := netio.NewTCPCtx(localA, true)
		if err != nil {
			t.Error(err.Error())
		}

		localCtxA.EncryptWrite([]byte(msgA))
		localCtxA.DecryptRead(bufA)
	}()

	go func() {
		listenerB, err := net.Listen("tcp", ":8888")
		if err != nil {
			t.Error(err.Error())
		}
		defer listenerB.Close()

		connB, err := listenerB.Accept()
		if err != nil {
			t.Error(err.Error())
		}
		defer connB.Close()

		connCtxB, err := netio.NewTCPCtx(connB, true)
		if err != nil {
			t.Error(err.Error())
		}

		connCtxB.EncryptWrite([]byte(msgB))
		connCtxB.DecryptRead(bufB)
	}()

	Local2Remote(":9999", "127.0.0.1:8888", true, true)
	if string(bufA[:len(msgB)]) != msgB || string(bufB[:len(msgA)]) != msgA {
		t.Error("Remote2Remote error")
	}
}
