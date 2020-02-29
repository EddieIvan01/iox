package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

var (
	SECRET_KEY []byte
	IV         []byte
)

func expand32(key []byte) ([]byte, []byte) {
	if len(key) >= 0x20 {
		return key[:0x10], key[0x10:0x20]
	}

	var c byte = 0x20 - byte(len(key)&0x1F)

	for i := 0; i < int(c); i++ {
		key = append(key, c)
	}
	return key[:0x10], key[0x10:0x20]
}

func ExpandKey(key []byte) {
	SECRET_KEY, IV = expand32(key)
}

type Cipher struct {
	c cipher.Stream
}

// AES-128-CTR
func NewCipherPair() (*Cipher, *Cipher, error) {
	blockA, err := aes.NewCipher(SECRET_KEY)
	if err != nil {
		return nil, nil, err
	}

	blockB, err := aes.NewCipher(SECRET_KEY)
	if err != nil {
		return nil, nil, err
	}

	return &Cipher{cipher.NewCTR(blockA, IV)},
		&Cipher{cipher.NewCTR(blockB, IV)},
		nil
}

func RandomIV() ([]byte, error) {
	iv := make([]byte, 0x10)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}
	return iv, nil
}

func NewCipher(iv []byte) (*Cipher, error) {
	block, err := aes.NewCipher(SECRET_KEY)
	if err != nil {
		return nil, err
	}

	return &Cipher{
		c: cipher.NewCTR(block, iv),
	}, nil
}

func (c Cipher) StreamXOR(dst []byte, src []byte) {
	c.c.XORKeyStream(dst, src)
}
