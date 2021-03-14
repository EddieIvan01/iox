package netio

import (
	"io"
	"iox/option"
	"sync"

	"github.com/klauspost/compress/s2"
)

var (
	TCPBufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, option.TCP_BUFFER_SIZE, option.TCP_BUFFER_SIZE)
		},
	}

	UDPBufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, option.UDP_PACKET_MAX_SIZE, option.UDP_PACKET_MAX_SIZE)
		},
	}

	compressReaderPool sync.Pool
	compressWriterPool sync.Pool
)

func GetCompressReader(r io.Reader) *s2.Reader {
	cr, ok := compressReaderPool.Get().(*s2.Reader)
	if !ok {
		return s2.NewReader(r)
	}

	cr.Reset(r)
	return cr
}

func PutCompressReader(cr *s2.Reader) {
	compressReaderPool.Put(cr)
}

func GetCompressWriter(w io.Writer) *s2.Writer {
	cw, ok := compressWriterPool.Get().(*s2.Writer)
	if !ok {
		return s2.NewWriter(w, s2.WriterBetterCompression())
	}

	cw.Reset(w)
	return cw
}

func PutCompressWriter(cw *s2.Writer) {
	compressWriterPool.Put(cw)
}
