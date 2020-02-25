package operate

import (
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
		// TODO
	}
}

func Local2Local(localA string, localB string, laenc bool, lbenc bool) {
	if option.PROTOCOL == "TCP" {
		logger.Success("Forward between %s and %s", localA, localB)

		for {
			signal := make(chan byte)
			var localConnA, localConnB net.Conn

			go func() {
				// define listener as a local variable
				// to control accepting connection timing
				listenerA, err := net.Listen("tcp", localA)
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
				listenerB, err := net.Listen("tcp", localB)
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
		// TODO
	}
}

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
		// TODO
	}
}
