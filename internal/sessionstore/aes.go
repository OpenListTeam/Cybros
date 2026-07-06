package sessionstore

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"cybros/consts"
)

func newSessionAEAD(key [32]byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorCreateSessionCipher, err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorCreateSessionGCM, err)
	}
	return aead, nil
}
