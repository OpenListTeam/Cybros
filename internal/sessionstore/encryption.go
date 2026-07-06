package sessionstore

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"cybros/consts"
)

const sessionAdditionalData = "cybros-session-v1"

type encryptedSessionFile struct {
	Salt  string `json:"salt"`
	Nonce string `json:"nonce"`
	Data  string `json:"data"`
}

func decodeSessionFile(data []byte, path string) (encryptedSessionFile, bool, error) {
	var file encryptedSessionFile
	err := json.Unmarshal(data, &file)
	if err != nil {
		return encryptedSessionFile{}, false, fmt.Errorf(consts.ErrorDecodeSessionFile, path, err)
	}

	if file.Salt == "" && file.Nonce == "" && file.Data == "" {
		return encryptedSessionFile{}, true, nil
	}
	if file.Salt == "" || file.Nonce == "" || file.Data == "" {
		return encryptedSessionFile{}, false, fmt.Errorf(consts.ErrorMalformedEncryptedSessionFile, path)
	}
	return file, false, nil
}

func decryptSessionData(passphrase string, file encryptedSessionFile, path string) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(file.Salt)
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorDecodeSessionSalt, err)
	}
	if len(salt) != sessionSaltSize {
		return nil, fmt.Errorf(consts.ErrorDecodeSessionSaltLength, len(salt))
	}
	nonce, err := base64.StdEncoding.DecodeString(file.Nonce)
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorDecodeSessionNonce, err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(file.Data)
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorDecodeSessionData, err)
	}

	key, err := deriveSessionKey(passphrase, salt, sessionKDFIterations)
	if err != nil {
		return nil, err
	}

	aead, err := newSessionAEAD(key)
	if err != nil {
		return nil, err
	}
	if len(nonce) != aead.NonceSize() {
		return nil, fmt.Errorf(consts.ErrorDecodeSessionNonceLength, len(nonce))
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, []byte(sessionAdditionalData))
	if err == nil {
		return plaintext, nil
	}
	return nil, fmt.Errorf(consts.ErrorDecryptSessionAuthentication, path)
}

func encryptSessionData(passphrase string, data []byte) ([]byte, error) {
	salt, err := randomBytes(sessionSaltSize, "session salt")
	if err != nil {
		return nil, err
	}

	key, err := deriveSessionKey(passphrase, salt, sessionKDFIterations)
	if err != nil {
		return nil, err
	}

	aead, err := newSessionAEAD(key)
	if err != nil {
		return nil, err
	}

	nonce, err := randomBytes(aead.NonceSize(), "session nonce")
	if err != nil {
		return nil, err
	}

	file := encryptedSessionFile{
		Salt:  base64.StdEncoding.EncodeToString(salt),
		Nonce: base64.StdEncoding.EncodeToString(nonce),
		Data:  base64.StdEncoding.EncodeToString(aead.Seal(nil, nonce, data, []byte(sessionAdditionalData))),
	}

	out, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorEncodeEncryptedSessionFile, err)
	}
	return append(out, '\n'), nil
}

func randomBytes(size int, label string) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		return nil, fmt.Errorf(consts.ErrorGenerateRandomBytes, label, err)
	}
	return data, nil
}
