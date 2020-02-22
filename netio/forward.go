package netio

import (
	"io"
	"iox/logger"
	"iox/option"
)

// Memory optimized
func CipherCopy(dst Ctx, src Ctx) (int64, error) {
	buffer := make([]byte, option.BUFFER_SIZE)
	var written int64
	var err error

	for {
		var nr int
		var er error
		if src.IsEncrypted() {
			nr, er = src.DecryptRead(buffer)
		} else {
			nr, er = src.Read(buffer)
		}

		if nr > 0 {
			logger.Info(" <== [%d bytes] ==", nr)

			var nw int
			var ew error
			if dst.IsEncrypted() {
				nw, ew = dst.EncryptWrite(buffer[:nr])
			} else {
				nw, ew = dst.Write(buffer[:nr])
			}

			if nw > 0 {
				logger.Info(" == [%d bytes] ==> ", nw)
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
		CipherCopy(ctxA, ctxB)
		signal <- struct{}{}
	}()

	go func() {
		CipherCopy(ctxB, ctxA)
		signal <- struct{}{}
	}()

	<-signal
}
