package crypto

import (
	"crypto/aes"
	"crypto/cipher"
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

type Cipher struct {
	c cipher.Stream
}

// AES-128-CTR
func NewCipherPair(key []byte) (*Cipher, *Cipher, error) {
	secretKey, iv := expand32(key)
	blockA, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, nil, err
	}

	blockB, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, nil, err
	}

	return &Cipher{cipher.NewCTR(blockA, iv)},
		&Cipher{cipher.NewCTR(blockB, iv)},
		nil
}

func (c Cipher) StreamXOR(dst []byte, src []byte) {
	c.c.XORKeyStream(dst, src)
}
