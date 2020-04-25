package operate

import (
	"iox/crypto"
	"iox/logger"
	"iox/netio"
	"iox/option"
	"net"
	"time"
)

// local is :port
// remote is ip:port
// Local2Remote(":9999", "1.1.1.1:9999")
func Local2Remote(local string, remote string, lenc bool, renc bool) {
	if option.PROTOCOL == "TCP" {
		listener, err := net.Listen("tcp", local)
		if err != nil {
			logger.Warn(
				"Listen on %s error: %s",
				local, err.Error(),
			)
			return
		}
		defer listener.Close()

		logger.Success("Forward between %s and %s", local, remote)

		for {
			logger.Info("Wait for connection on %s", local)

			localConn, err := listener.Accept()
			if err != nil {
				logger.Warn(
					"Handle local connect error: %s",
					err.Error(),
				)
				continue
			}

			logger.Info(
				"Connection from %s",
				localConn.RemoteAddr().String(),
			)
			logger.Info("Connecting " + remote)

			go func() {
				defer localConn.Close()

				localConnCtx, err := netio.NewTCPCtx(localConn, lenc)
				if err != nil {
					logger.Warn(
						"Handle local connect error: %s",
						err.Error(),
					)
					return
				}

				remoteConn, err := net.DialTimeout(
					"tcp",
					remote,
					time.Millisecond*time.Duration(option.TIMEOUT),
				)
				if err != nil {
					logger.Warn("Connect remote %s error: %s",
						remote, err.Error(),
					)
					return
				}
				defer remoteConn.Close()

				remoteConnCtx, err := netio.NewTCPCtx(remoteConn, renc)
				if err != nil {
					logger.Warn("Connect remote %s error: %s",
						remote, err.Error(),
					)
					return
				}

				logger.Info(
					"Open pipe: %s <== FWD ==> %s",
					localConn.RemoteAddr().String(),
					remoteConn.RemoteAddr().String(),
				)

				netio.PipeForward(localConnCtx, remoteConnCtx)

				logger.Info(
					"Close pipe: %s <== FWD ==> %s",
					localConn.RemoteAddr().String(),
					remoteConn.RemoteAddr().String(),
				)
			}()
		}

	} else {
		localAddr, err := net.ResolveUDPAddr("udp", local)
		if err != nil {
			logger.Warn(
				"Parse udp address %s error: %s",
				local, err.Error(),
			)
			return
		}
		listener, err := net.ListenUDP("udp", localAddr)
		if err != nil {
			logger.Warn(
				"Listen udp on %s error: %s",
				local, err.Error(),
			)
			return
		}

		remoteAddr, err := net.ResolveUDPAddr("udp", remote)
		if err != nil {
			logger.Warn(
				"Parse udp address %s error: %s",
				local, err.Error(),
			)
			return
		}
		remoteConn, err := net.DialUDP("udp", nil, remoteAddr)
		if err != nil {
			logger.Warn(
				"Dial remote udp %s error: %s",
				local, err.Error(),
			)
			return
		}

		listenerCtx, err := netio.NewUDPCtx(listener, lenc, false)
		if err != nil {
			return
		}
		remoteCtx, err := netio.NewUDPCtx(remoteConn, renc, true)
		if err != nil {
			return
		}

		logger.Success("Forward udp between %s and %s", local, remote)
		netio.ForwardUDP(listenerCtx, remoteCtx)
	}
}

func Local2Local(localA string, localB string, laenc bool, lbenc bool) {
	if option.PROTOCOL == "TCP" {
		logger.Success("Forward between %s and %s", localA, localB)

		var listenerA net.Listener
		var listenerB net.Listener

		for {
			signal := make(chan byte)
			var localConnA, localConnB net.Conn

			go func() {
				// Call listener.Close when goroutine returns.
				// Listener in Go will release the port immediately
				// after calling listener.Close without waiting for TIME_WAIT
				var err error
				listenerA, err = net.Listen("tcp", localA)
				if err != nil {
					logger.Warn(
						"Listen on %s error: %s",
						localA, err.Error(),
					)
					return
				}
				defer listenerA.Close()

				for {
					logger.Info(
						"Wait for connection on %s",
						localA,
					)

					var err error
					localConnA, err = listenerA.Accept()
					if err != nil {
						logger.Warn(
							"Handle connection error: %s",
							err.Error(),
						)
						continue
					}
					break
				}
				signal <- 'A'
			}()

			go func() {
				var err error
				listenerB, err = net.Listen("tcp", localB)
				if err != nil {
					logger.Warn(
						"Listen on %s error: %s",
						localB, err.Error(),
					)
					return
				}
				defer listenerB.Close()

				for {
					logger.Info(
						"Wait for connection on %s",
						localB,
					)

					var err error
					localConnB, err = listenerB.Accept()
					if err != nil {
						logger.Warn(
							"Handle connection error: %s",
							err.Error(),
						)
						continue
					}
					break
				}
				signal <- 'B'
			}()

			switch <-signal {
			case 'A':
				logger.Info(
					"%s connected, waiting for %s",
					localA, localB,
				)
			case 'B':
				logger.Info(
					"%s connected, waiting for %s",
					localB, localA,
				)
			}

			<-signal

			go func() {
				defer localConnA.Close()
				defer localConnB.Close()

				localConnCtxA, err := netio.NewTCPCtx(localConnA, laenc)
				if err != nil {
					logger.Warn(
						"handle local %s error: %s",
						localA, err.Error(),
					)
				}

				localConnCtxB, err := netio.NewTCPCtx(localConnB, lbenc)
				if err != nil {
					logger.Warn(
						"handle local %s error: %s",
						localB, err.Error(),
					)
				}

				logger.Info(
					"Open pipe: %s <== FWD ==> %s",
					localConnA.RemoteAddr().String(),
					localConnB.RemoteAddr().String(),
				)
				netio.PipeForward(localConnCtxA, localConnCtxB)
				logger.Info(
					"Close pipe: %s <== FWD ==> %s",
					localConnA.RemoteAddr().String(),
					localConnB.RemoteAddr().String(),
				)
			}()
		}
	} else {
		localAddrA, err := net.ResolveUDPAddr("udp", localA)
		if err != nil {
			logger.Warn(
				"Parse udp address %s error: %s",
				localA, err.Error(),
			)
			return
		}
		listenerA, err := net.ListenUDP("udp", localAddrA)
		if err != nil {
			logger.Warn(
				"Listen udp on %s error: %s",
				localA, err.Error(),
			)
			return
		}
		localAddrB, err := net.ResolveUDPAddr("udp", localB)
		if err != nil {
			logger.Warn(
				"Parse udp address %s error: %s",
				localB, err.Error(),
			)
			return
		}
		listenerB, err := net.ListenUDP("udp", localAddrB)
		if err != nil {
			logger.Warn(
				"Listen udp on %s error: %s",
				localB, err.Error(),
			)
			return
		}

		listenerCtxA, err := netio.NewUDPCtx(listenerA, laenc, false)
		if err != nil {
			return
		}
		listenerCtxB, err := netio.NewUDPCtx(listenerB, lbenc, false)
		if err != nil {
			return
		}

		logger.Success("Forward udp between %s and %s", localA, localB)
		netio.ForwardUnconnectedUDP(listenerCtxA, listenerCtxB)
	}
}

// When you make a multistage UDP connection, this function must be called last
func Remote2Remote(remoteA string, remoteB string, raenc bool, rbenc bool) {
	if option.PROTOCOL == "TCP" {
		logger.Success("Forward between %s and %s", remoteA, remoteB)

		for {
			var remoteConnA net.Conn
			var remoteConnB net.Conn

			signal := make(chan struct{})

			go func() {
				for {
					var err error
					logger.Info(
						"Connecting remote %s",
						remoteA,
					)

					remoteConnA, err = net.DialTimeout(
						"tcp",
						remoteA,
						time.Millisecond*time.Duration(option.TIMEOUT),
					)
					if err != nil {
						logger.Info(
							"Connect remote %s error, retrying",
							remoteA,
						)
						time.Sleep(1500 * time.Millisecond)
						continue
					}

					break
				}

				signal <- struct{}{}
			}()

			go func() {
				for {
					var err error
					logger.Info(
						"Connecting remote %s",
						remoteB,
					)

					remoteConnB, err = net.DialTimeout(
						"tcp",
						remoteB,
						time.Millisecond*time.Duration(option.TIMEOUT),
					)
					if err != nil {
						logger.Info(
							"Connect remote %s error, retrying",
							remoteB,
						)
						time.Sleep(1500 * time.Millisecond)
						continue
					}
					break
				}

				signal <- struct{}{}
			}()

			<-signal
			<-signal

			go func() {
				defer func() {
					if remoteConnA != nil {
						remoteConnA.Close()
					}

					if remoteConnB != nil {
						remoteConnB.Close()
					}
				}()

				if remoteConnA != nil && remoteConnB != nil {
					remoteConnCtxA, err := netio.NewTCPCtx(remoteConnA, raenc)
					if err != nil {
						logger.Warn(
							"Handle remote %s error: %s",
							remoteA, err.Error(),
						)
					}
					remoteConnCtxB, err := netio.NewTCPCtx(remoteConnB, rbenc)
					if err != nil {
						logger.Warn(
							"Handle remote %s error: %s",
							remoteB, err.Error(),
						)
					}

					logger.Info(
						"Start pipe: %s <== FWD ==> %s",
						remoteConnA.RemoteAddr().String(),
						remoteConnB.RemoteAddr().String(),
					)
					netio.PipeForward(remoteConnCtxA, remoteConnCtxB)
					logger.Info(
						"Close pipe: %s <== FWD ==> %s",
						remoteConnA.RemoteAddr().String(),
						remoteConnB.RemoteAddr().String(),
					)
				}
			}()
		}
	} else {
		remoteAddrA, err := net.ResolveUDPAddr("udp", remoteA)
		if err != nil {
			logger.Warn(
				"Parse udp address %s error: %s",
				remoteA, err.Error(),
			)
			return
		}
		remoteConnA, err := net.DialUDP("udp", nil, remoteAddrA)
		if err != nil {
			logger.Warn(
				"Dial remote udp %s error: %s",
				remoteA, err.Error(),
			)
			return
		}
		remoteAddrB, err := net.ResolveUDPAddr("udp", remoteB)
		if err != nil {
			logger.Warn(
				"Parse udp address %s error: %s",
				remoteB, err.Error(),
			)
			return
		}
		remoteConnB, err := net.DialUDP("udp", nil, remoteAddrB)
		if err != nil {
			logger.Warn(
				"Dial remote udp %s error: %s",
				remoteB, err.Error(),
			)
			return
		}

		remoteCtxA, err := netio.NewUDPCtx(remoteConnA, raenc, true)
		if err != nil {
			return
		}
		remoteCtxB, err := netio.NewUDPCtx(remoteConnB, rbenc, true)
		if err != nil {
			return
		}

		// I need to send init packet to register the remote address
		// Even tough target is not `iox`, it doesn't matter
		//
		// There is a design fault here, and I need to consider
		// the case where the FORWARD_WITHOUT_DEC flag
		// is set but actually needs to be encrypted,
		// otherwise there is no IV in the ciphertext,
		// the opposite cannot process it
		if raenc {
			iv, err := crypto.RandomNonce()
			cipher, err := crypto.NewCipher(iv)
			if err != nil {
				return
			}

			b := make([]byte, 4, 20)
			copy(b, netio.UDP_INIT_PACKET)

			cipher.StreamXOR(b, b)
			b = append(b, iv...)
			remoteCtxA.Write(b)

		} else {
			remoteCtxA.Write(netio.UDP_INIT_PACKET)
		}
		if rbenc {
			iv, err := crypto.RandomNonce()
			cipher, err := crypto.NewCipher(iv)
			if err != nil {
				return
			}

			b := make([]byte, 4, 20)
			copy(b, netio.UDP_INIT_PACKET)

			cipher.StreamXOR(b, b)
			b = append(b, iv...)
			remoteCtxB.Write(b)

		} else {
			remoteCtxB.Write(netio.UDP_INIT_PACKET)
		}

		logger.Success("Forward udp between %s and %s", remoteA, remoteB)
		netio.ForwardUDP(remoteCtxA, remoteCtxB)
	}
}
