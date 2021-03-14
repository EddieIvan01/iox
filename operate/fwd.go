package operate

import (
	"iox/logger"
	"iox/netio"
	"iox/option"
	"net"
	"os"
	"time"
)

func local2RemoteReliableProto(localDesc *option.SocketDesc, remoteDesc *option.SocketDesc) {
	listener, err := localDesc.GetListener()
	if err != nil {
		logger.Warn("Listen on %s error: %s", localDesc.Addr, err.Error())
		return
	}
	defer listener.Close()

	for {
		logger.Info("Wait for connection on %s", localDesc.Addr)

		localConn, err := listener.Accept()
		if err != nil {
			if _, ok := err.(*net.OpError); ok || err == net.ErrClosed {
				logger.Success("Smux session has been closed")
				os.Exit(0)
			}
			logger.Warn("Handle local connect error: %s", err.Error())
			continue
		}

		go func() {
			logger.Info("Connection from %s", localConn.RemoteAddr().String())
			logger.Info("Connecting " + remoteDesc.Addr)

			localConnCtx, err := netio.NewTCPCtx(localConn, localDesc.Secret, localDesc.Compress)
			defer localConnCtx.Close()
			if err != nil {
				logger.Warn("Handle local connect error: %s", err.Error())
				return
			}

			remoteConn, err := remoteDesc.GetConn()
			if err != nil {
				logger.Warn("Connect remote %s error: %s", remoteDesc.Addr, err.Error())
				return
			}

			remoteConnCtx, err := netio.NewTCPCtx(remoteConn, remoteDesc.Secret, remoteDesc.Compress)
			defer remoteConnCtx.Close()
			if err != nil {
				logger.Warn("Connect remote %s error: %s", remoteDesc.Addr, err.Error())
				return
			}

			logger.Info("Open pipe: %s <== FWD ==> %s",
				localConn.RemoteAddr().String(), remoteConn.RemoteAddr().String())
			netio.PipeForward(localConnCtx, remoteConnCtx)
			logger.Info("Close pipe: %s <== FWD ==> %s",
				localConn.RemoteAddr().String(), remoteConn.RemoteAddr().String())
		}()
	}

}

func local2RemoteUDP(localDesc *option.SocketDesc, remoteDesc *option.SocketDesc) {
	listener, err := localDesc.GetUDPConn()
	if err != nil {
		logger.Warn("Listen udp on %s error: %s", localDesc.Addr, err.Error())
		return
	}
	defer listener.Close()

	remoteConn, err := remoteDesc.GetUDPConn()
	if err != nil {
		logger.Warn("Dial remote udp %s error: %s", remoteDesc.Addr, err.Error())
		return
	}
	defer remoteConn.Close()

	listenerCtx, err := netio.NewUDPCtx(listener, localDesc.Secret, false)
	if err != nil {
		return
	}
	remoteCtx, err := netio.NewUDPCtx(remoteConn, remoteDesc.Secret, true)
	if err != nil {
		return
	}

	netio.ForwardUDP(listenerCtx, remoteCtx)
}

func Local2Remote(localDesc *option.SocketDesc, remoteDesc *option.SocketDesc) {
	logger.Success("Forward traffic between %s and %s",
		localDesc, remoteDesc)

	if localDesc.IsProtoReliable() {
		local2RemoteReliableProto(localDesc, remoteDesc)
	} else {
		local2RemoteUDP(localDesc, remoteDesc)
	}
}

func local2LocalReliableProto(localDescA *option.SocketDesc, localDescB *option.SocketDesc) {
	var listenerA net.Listener
	var listenerB net.Listener

	signal := make(chan struct{}, 1)
	go func() {
		var err error
		listenerA, err = localDescA.GetListener()
		if err != nil {
			logger.Warn("Listen on %s error: %s", localDescA.Addr, err.Error())
			return
		}
		signal <- struct{}{}
	}()
	go func() {
		var err error
		listenerB, err = localDescB.GetListener()
		if err != nil {
			logger.Warn("Listen on %s error: %s", localDescB.Addr, err.Error())
			return
		}
		signal <- struct{}{}
	}()
	<-signal
	<-signal

	defer listenerA.Close()
	defer listenerB.Close()

	for {
		var localConnA net.Conn
		var localConnB net.Conn

		go func() {
			for {
				logger.Info("Wait for connection on %s", localDescA.Addr)

				var err error
				localConnA, err = listenerA.Accept()
				if err != nil {
					if _, ok := err.(*net.OpError); ok || err == net.ErrClosed {
						logger.Success("Smux session has been closed")
						os.Exit(0)
					}
					logger.Warn("Handle connection error: %s", err.Error())
					continue
				}
				break
			}
			signal <- struct{}{}
		}()

		go func() {
			for {
				logger.Info("Wait for connection on %s", localDescB.Addr)

				var err error
				localConnB, err = listenerB.Accept()
				if err != nil {
					if _, ok := err.(*net.OpError); ok || err == net.ErrClosed {
						logger.Success("Smux session has been closed")
						os.Exit(0)
					}
					logger.Warn("Handle connection error: %s", err.Error())
					continue
				}
				break
			}
			signal <- struct{}{}
		}()

		<-signal
		<-signal

		go func() {
			localConnCtxA, err := netio.NewTCPCtx(localConnA, localDescA.Secret, localDescA.Compress)
			defer localConnCtxA.Close()
			if err != nil {
				return
			}

			localConnCtxB, err := netio.NewTCPCtx(localConnB, localDescB.Secret, localDescB.Compress)
			defer localConnCtxB.Close()
			if err != nil {
				return
			}

			logger.Info("Open pipe: %s <== FWD ==> %s",
				localConnA.RemoteAddr().String(), localConnB.RemoteAddr().String())
			netio.PipeForward(localConnCtxA, localConnCtxB)
			logger.Info("Close pipe: %s <== FWD ==> %s",
				localConnA.RemoteAddr().String(), localConnB.RemoteAddr().String())
		}()
	}
}

func local2LocalUDP(localDescA *option.SocketDesc, localDescB *option.SocketDesc) {
	listenerA, err := localDescA.GetUDPConn()
	if err != nil {
		logger.Warn("Listen udp on %s error: %s", localDescA.Addr, err.Error())
		return
	}
	defer listenerA.Close()

	listenerB, err := localDescB.GetUDPConn()
	if err != nil {
		logger.Warn("Listen udp on %s error: %s", localDescB.Addr, err.Error())
		return
	}
	defer listenerB.Close()

	listenerCtxA, err := netio.NewUDPCtx(listenerA, localDescA.Secret, false)
	if err != nil {
		return
	}
	listenerCtxB, err := netio.NewUDPCtx(listenerB, localDescB.Secret, false)
	if err != nil {
		return
	}

	netio.ForwardUnconnectedUDP(listenerCtxA, listenerCtxB)
}

func Local2Local(localDescA *option.SocketDesc, localDescB *option.SocketDesc) {
	logger.Success("Forward traffic between %s and %s",
		localDescA, localDescB)

	if localDescA.IsProtoReliable() {
		local2LocalReliableProto(localDescA, localDescB)
	} else {
		local2LocalUDP(localDescA, localDescB)
	}
}

func remote2RemoteReliableProto(remoteDescA *option.SocketDesc, remoteDescB *option.SocketDesc) {
	for {
		var remoteConnA net.Conn
		var remoteConnB net.Conn

		signal := make(chan struct{})

		go func() {
			for {
				var err error
				logger.Info("Connecting remote %s", remoteDescA.Addr)

				remoteConnA, err = remoteDescA.GetConn()
				if err != nil {
					logger.Info("Connect remote %s error, retrying", remoteDescA.Addr)
					time.Sleep(option.CONNECTING_RETRY_DURATION * time.Millisecond)
					continue
				}
				break
			}

			signal <- struct{}{}
		}()

		go func() {
			for {
				var err error
				logger.Info("Connecting remote %s", remoteDescB.Addr)

				remoteConnB, err = remoteDescB.GetConn()
				if err != nil {
					logger.Info("Connect remote %s error, retrying", remoteDescB.Addr)
					time.Sleep(option.CONNECTING_RETRY_DURATION * time.Millisecond)
					continue
				}
				break
			}

			signal <- struct{}{}
		}()

		<-signal
		<-signal

		go func() {
			remoteConnCtxA, err := netio.NewTCPCtx(remoteConnA, remoteDescA.Secret, remoteDescA.Compress)
			defer remoteConnCtxA.Close()
			if err != nil {
				return
			}
			remoteConnCtxB, err := netio.NewTCPCtx(remoteConnB, remoteDescB.Secret, remoteDescB.Compress)
			defer remoteConnCtxB.Close()
			if err != nil {
				return
			}

			logger.Info("Start pipe: %s <== FWD ==> %s",
				remoteConnA.RemoteAddr().String(), remoteConnB.RemoteAddr().String())
			netio.PipeForward(remoteConnCtxA, remoteConnCtxB)
			logger.Info("Close pipe: %s <== FWD ==> %s",
				remoteConnA.RemoteAddr().String(), remoteConnB.RemoteAddr().String())

		}()
	}
}

func remote2RemoteUDP(remoteDescA *option.SocketDesc, remoteDescB *option.SocketDesc) {
	remoteConnA, err := remoteDescA.GetUDPConn()
	if err != nil {
		logger.Warn("Dial remote udp %s error: %s", remoteDescA.Addr, err.Error())
		return
	}
	defer remoteConnA.Close()

	remoteConnB, err := remoteDescB.GetUDPConn()
	if err != nil {
		logger.Warn("Dial remote udp %s error: %s", remoteDescB.Addr, err.Error())
		return
	}
	defer remoteConnB.Close()

	remoteCtxA, err := netio.NewUDPCtx(remoteConnA, remoteDescA.Secret, true)
	if err != nil {
		return
	}
	remoteCtxB, err := netio.NewUDPCtx(remoteConnB, remoteDescB.Secret, true)
	if err != nil {
		return
	}

	remoteCtxA.Write(netio.UDP_INIT_PACKET)
	remoteCtxB.Write(netio.UDP_INIT_PACKET)

	netio.ForwardUDP(remoteCtxA, remoteCtxB)
}

func Remote2Remote(remoteDescA *option.SocketDesc, remoteDescB *option.SocketDesc) {
	logger.Success("Forward traffic between %s and %s",
		remoteDescA, remoteDescB)

	if remoteDescA.IsProtoReliable() {
		remote2RemoteReliableProto(remoteDescA, remoteDescB)
	} else {
		remote2RemoteUDP(remoteDescA, remoteDescB)
	}
}
