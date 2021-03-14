/*
 Third-party chacha20 lib from https://github.com/Yawning/chacha20
 Its module is wrong so we couldn't import by go mod
*/
package crypto

import (
	"crypto/rand"
	"io"

	"iox/crypto/chacha20"
)

var (
	SECRET_KEY []byte
)

func shuffle(bs []byte) {
	for i := range bs {
		bs[i] ^= byte(i) ^ bs[(i+1)%len(bs)]*((bs[len(bs)-1-i]*bs[i])%255)
	}
}

func ExpandKey(key []byte) {
	SECRET_KEY = make([]byte, 0x20)

	if len(key) < 0x20 {
		var c byte = 0x20 - byte(len(key)&0x1f)

		for i := 0; i < int(c); i++ {
			key = append(key, c)
		}
	}

	copy(SECRET_KEY, key[:0x20])

	for i := range SECRET_KEY {
		SECRET_KEY[i] = (SECRET_KEY[i] + byte(i)%255)
	}

	shuffle(SECRET_KEY)
}

type Cipher struct {
	c *chacha20.Cipher
}

func RandomNonce() ([]byte, error) {
	iv := make([]byte, 0x18)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}
	return iv, nil
}

func NewCipher(nonce []byte) (*Cipher, error) {
	cc, err := chacha20.New(SECRET_KEY, nonce)
	if err != nil {
		return nil, err
	}

	return &Cipher{
		c: cc,
	}, nil
}

func (c Cipher) StreamXOR(dst []byte, src []byte) {
	c.c.XORKeyStream(dst, src)
}

type Reader struct {
	r      io.Reader
	cipher *Cipher
}

type Writer struct {
	w      io.Writer
	cipher *Cipher
}

func NewReader(r io.Reader, iv []byte) (*Reader, error) {
	cipher, err := NewCipher(iv)
	if err != nil {
		return nil, err
	}

	return &Reader{
		r:      r,
		cipher: cipher,
	}, nil
}

func NewWriter(w io.Writer, iv []byte) (*Writer, error) {
	cipher, err := NewCipher(iv)
	if err != nil {
		return nil, err
	}

	return &Writer{
		w:      w,
		cipher: cipher,
	}, nil
}

func (r *Reader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	if err != nil {
		return n, err
	}

	r.cipher.StreamXOR(b[:n], b[:n])
	return n, nil
}

func (w *Writer) Write(b []byte) (int, error) {
	w.cipher.StreamXOR(b, b)
	n, err := w.w.Write(b)
	return n, err
}
