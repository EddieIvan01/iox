package netio

import (
	"io"
	"iox/logger"
	"iox/option"
	"reflect"
	"unsafe"
)

func Copy(dst Ctx, src Ctx) (int64, error) {
	buffer := TCPBufferPool.Get().([]byte)
	defer TCPBufferPool.Put(buffer)

	var written int64
	var err error

	for {
		var nr int
		var er error

		nr, er = src.Read(buffer)

		if nr > 0 {
			var nw int
			var ew error

			nw, ew = dst.Write(buffer[:nr])

			if nw > 0 {
				logger.Info("<== [%d bytes] ==> ", nw)
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return written, err
}

func PipeForward(ctxA Ctx, ctxB Ctx) {
	signal := make(chan struct{}, 1)

	go func() {
		Copy(ctxA, ctxB)
		signal <- struct{}{}
	}()

	go func() {
		Copy(ctxB, ctxA)
		signal <- struct{}{}
	}()

	<-signal
}

var UDP_INIT_PACKET = []byte{
	0xcc, 0xdd, 0xee, 0xff,
}

func isUDPInitPacket(buf []byte) bool {
	return buf[0] == UDP_INIT_PACKET[0] && buf[1] == UDP_INIT_PACKET[1] &&
		buf[2] == UDP_INIT_PACKET[2] && buf[3] == UDP_INIT_PACKET[3]
}

// This function will run forever
// If need to do performance optimization in future, I will consider a go-routine pool here,
// but it will introduce the mutex-lock overhead
func ForwardUDP(ctxA Ctx, ctxB Ctx) {
	go func() {
		buffer := UDPBufferPool.Get().([]byte)
		defer UDPBufferPool.Put(buffer)
		for {
			nr, _ := ctxA.Read(buffer)
			if nr > 0 {
				if nr == 4 && isUDPInitPacket(buffer[:4]) {
					continue
				}

				nw, _ := ctxB.Write(buffer[:nr])
				if nw > 0 {
					logger.Info("<== [%d bytes] ==>", nw)
				}
			}
		}
	}()

	go func() {
		buffer := UDPBufferPool.Get().([]byte)
		defer UDPBufferPool.Put(buffer)
		for {
			nr, _ := ctxB.Read(buffer)
			if nr > 0 {
				if nr == 4 && isUDPInitPacket(buffer[:4]) {
					continue
				}

				nw, _ := ctxA.Write(buffer[:nr])
				if nw > 0 {
					logger.Info("<== [%d bytes] ==>", nw)
				}
			}
		}
	}()

	select {}
}

func hackResizeLength(bufp *[]byte) {
	(*(*reflect.SliceHeader)(unsafe.Pointer(bufp))).Len = option.UDP_PACKET_MAX_SIZE
}

// Each socket only writes the packet to the address which last sent packet to self recently,
// instead of broadcasting to all address
func ForwardUnconnectedUDP(ctxA Ctx, ctxB Ctx) {
	addrRegistedA := false
	addrRegistedB := false
	addrRegistedSignalA := make(chan struct{})
	addrRegistedSignalB := make(chan struct{})

	packetChannelA := make(chan []byte, option.UDP_PACKET_CHANNEL_SIZE)
	packetChannelB := make(chan []byte, option.UDP_PACKET_CHANNEL_SIZE)

	// A read
	go func() {
		for {
			buffer := UDPBufferPool.Get().([]byte)
			nr, _ := ctxA.Read(buffer)
			if nr > 0 {
				if !addrRegistedA {
					addrRegistedA = true
					addrRegistedSignalA <- struct{}{}
				}

				if !(nr == 4 && isUDPInitPacket(buffer[:4])) {
					packetChannelB <- buffer[:nr]
				}
			}
		}
	}()

	// B read
	go func() {
		for {
			buffer := UDPBufferPool.Get().([]byte)
			nr, _ := ctxB.Read(buffer)
			if nr > 0 {
				if !addrRegistedB {
					addrRegistedB = true
					addrRegistedSignalB <- struct{}{}
				}

				if !(nr == 4 && isUDPInitPacket(buffer[:4])) {
					packetChannelA <- buffer[:nr]
				}
			}
		}
	}()

	// A write
	go func() {
		<-addrRegistedSignalA
		var n int
		for {
			packet := <-packetChannelA
			n, _ = ctxA.Write(packet)
			hackResizeLength(&packet)
			UDPBufferPool.Put(packet)

			if n > 0 {
				logger.Info("<== [%d bytes] ==>", n)
			}
		}
	}()

	// B write
	go func() {
		<-addrRegistedSignalB
		var n int
		for {
			packet := <-packetChannelB
			n, _ = ctxB.Write(packet)
			hackResizeLength(&packet)
			UDPBufferPool.Put(packet)

			if n > 0 {
				logger.Info("<== [%d bytes] ==>", n)
			}
		}
	}()

	select {}
}
