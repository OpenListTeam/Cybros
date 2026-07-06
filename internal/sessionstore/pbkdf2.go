package sessionstore

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"fmt"

	"cybros/consts"
)

const (
	sessionKDFIterations = 65535
	sessionKeySize       = 32
	sessionSaltSize      = 32
)

func deriveSessionKey(passphrase string, salt []byte, iterations int) ([32]byte, error) {
	key, err := pbkdf2.Key(sha256.New, passphrase, salt, iterations, sessionKeySize)
	if err != nil {
		return [32]byte{}, fmt.Errorf(consts.ErrorDeriveSessionKey, err)
	}
	if len(key) != sessionKeySize {
		return [32]byte{}, fmt.Errorf(consts.ErrorDeriveSessionKeyLength, len(key))
	}

	var out [32]byte
	copy(out[:], key)
	return out, nil
}
