package operate

import (
	"iox/logger"
	"iox/netio"
	"iox/option"
	"iox/socks5"
	"net"
	"os"
	"os/signal"
)

func ProxyLocal(localDesc *option.SocketDesc) {
	listener, err := localDesc.GetListener()
	if err != nil {
		logger.Warn("Socks5 listen on %s error: %s", localDesc.Addr, err.Error())
		return
	}

	logger.Success("Start socks5 server on %s", localDesc.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if _, ok := err.(*net.OpError); ok || err == net.ErrClosed {
				logger.Success("Smux session has been closed")
				os.Exit(0)
			}
			logger.Warn("Socks5 handle local connect error: %s", err.Error())
			continue
		}

		go func() {
			connCtx, err := netio.NewTCPCtx(conn, localDesc.Secret, localDesc.Compress)
			defer connCtx.Close()
			if err != nil {
				return
			}

			socks5.HandleConnection(connCtx)
		}()
	}
}

func ProxyRemote(remoteDesc *option.SocketDesc) {
	ctlStreamCtx, err := clientHandshake(remoteDesc)
	if err != nil {
		logger.Warn(err.Error())
		return
	}
	defer ctlStreamCtx.Close()

	logger.Success("Remote socks5 handshake OK")

	connectRequest := make(chan uint8, MAX_CONNECTION)
	defer close(connectRequest)
	endSignal := make(chan struct{})

	// handle ctrl+C
	{
		sigs := make(chan os.Signal)
		signal.Notify(sigs, os.Interrupt)
		go func() {
			<-sigs
			ctlStreamCtx.Write(marshal(Protocol{
				CMD: CTL_CLEANUP,
				N:   0,
			}))
			logger.Success("Recv Ctrl+C, exit now")
			os.Exit(0)
		}()
	}

	// handle ctl stream
	go func() {
		for {
			pb, err := readUntilEnd(ctlStreamCtx)
			if err != nil {
				logger.Warn("Control connection has been closed, exit now")
				os.Exit(-1)
			}

			p := unmarshal(pb)
			switch p.CMD {
			case CTL_CONNECT_ME:
				connectRequest <- p.N
			case CTL_CLEANUP:
				endSignal <- struct{}{}
				return
			}
		}
	}()

	// handle CONNECT_ME request
	for {
		select {
		case <-endSignal:
			logger.Success("Recv exit signal from remote, exit now")
			return
		case n := <-connectRequest:
			for n > 0 {
				go func() {
					conn, err := remoteDesc.GetConn()
					if err != nil {
						logger.Info(err.Error())
						return
					}

					// Init UDP unconnected session
					if remoteDesc.Proto == "kcp" {
						conn.Write([]byte{0})
					}

					connCtx, err := netio.NewTCPCtx(conn, remoteDesc.Secret, remoteDesc.Compress)
					defer connCtx.Close()
					if err != nil {
						return
					}
					socks5.HandleConnection(connCtx)
				}()
				n--
			}
		}
	}
}

func ProxyRemoteL2L(ctlDesc *option.SocketDesc, localDesc *option.SocketDesc) {
	ctlListener, err := ctlDesc.GetListener()
	if err != nil {
		logger.Warn("Listen on %s error", ctlDesc.Addr)
		return
	}
	defer ctlListener.Close()

	logger.Info("Listen on %s for reverse socks5", ctlDesc.Addr)

	localListener, err := localDesc.GetListener()
	if err != nil {
		logger.Warn("Listen on %s error", localDesc.Addr)
		return
	}
	defer localListener.Close()

	ctlStreamCtx, err := serverHandshake(ctlListener, ctlDesc.Secret, ctlDesc.Compress)
	defer ctlStreamCtx.Close()
	if err != nil {
		logger.Warn(err.Error())
		return
	}

	logger.Success("Reverse socks5 server handshake OK")
	logger.Success("Socks5 server is listening on %s", localDesc.Addr)

	// handle ctrl+C
	{
		sigs := make(chan os.Signal)
		signal.Notify(sigs, os.Interrupt)
		go func() {
			<-sigs
			ctlStreamCtx.Write(marshal(Protocol{
				CMD: CTL_CLEANUP,
				N:   0,
			}))
			logger.Success("Recv Ctrl+C, exit now")
			os.Exit(0)
		}()
	}

	localConnBuffer := make(chan net.Conn, MAX_CONNECTION)
	defer close(localConnBuffer)

	// handle ctl stream read
	go func() {
		for {
			pb, err := readUntilEnd(ctlStreamCtx)
			if err != nil {
				logger.Warn("Control connection has been closed, exit now")
				os.Exit(-1)
			}

			p := unmarshal(pb)
			switch p.CMD {
			case CTL_CLEANUP:
				logger.Success("Recv exit signal from remote, exit now")
				os.Exit(0)
			}
		}
	}()

	// handle local connection
	go func() {
		for {
			localConn, err := localListener.Accept()
			if err != nil {
				if _, ok := err.(*net.OpError); ok || err == net.ErrClosed {
					logger.Success("Smux session has been closed")
					os.Exit(0)
				}
				logger.Info(err.Error())
				continue
			}

			localConnBuffer <- localConn

			_, err = ctlStreamCtx.Write(marshal(Protocol{
				CMD: CTL_CONNECT_ME,
				N:   1,
			}))
			if err != nil {
				logger.Warn("Control connection has been closed, exit now")
				os.Exit(-1)
			}
		}
	}()

	for {
		remoteStream, err := ctlListener.Accept()
		if err != nil {
			if _, ok := err.(*net.OpError); ok || err == net.ErrClosed {
				logger.Success("Smux session has been closed")
				os.Exit(0)
			}
			continue
		}

		// Init UDP unconnected session
		if ctlDesc.Proto == "kcp" {
			remoteStream.Read([]byte{0})
		}

		localConn := <-localConnBuffer

		go func() {
			remoteConnCtx, err := netio.NewTCPCtx(remoteStream, ctlDesc.Secret, ctlDesc.Compress)
			defer remoteConnCtx.Close()
			if err != nil {
				return
			}

			localConnCtx, err := netio.NewTCPCtx(localConn, localDesc.Secret, localDesc.Compress)
			defer localConnCtx.Close()
			if err != nil {
				return
			}

			netio.PipeForward(remoteConnCtx, localConnCtx)
		}()
	}
}
