package tgclient

import (
	"encoding/pem"
	"fmt"
	"time"

	"cybros/consts"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/telegram"
)

const telegramPublicKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA6LszBcC1LGzyr992NzE0ieY+BSaOW622Aa9Bd4ZHLl+TuFQ4lo4g
5nKaMBwK/BIb9xUfg0Q29/2mgIR6Zr9krM7HjuIcCzFvDtr+L0GQjae9H0pRB2OO
62cECs5HKhT5DZ98K33vmWiLowc621dQuwKWSQKjWf50XYFw42h21P2KXUGyp2y/
+aEyZ+uVgLLQbRA1dEjSDZ2iGRy12Mk5gpYc397aYp438fsJoHIgJ2lgMv5h7WY9
t6N/byY9Nw9p21Og3AoXSL2q/2IJ1WRUhebgAdGVMlV1fkuOQoEzR7EdpqtQD9Cs
5+bfo3Nhmcyvk5ftB0WkJ9z6bNZ7yxrP8wIDAQAB
-----END RSA PUBLIC KEY-----`

func NewOptions(storage telegram.SessionStorage, handler telegram.UpdateHandler) (telegram.Options, error) {
	publicKey, err := parsePublicKey()
	if err != nil {
		return telegram.Options{}, err
	}
	if handler == nil {
		handler = updateHandler{}
	}

	return telegram.Options{
		PublicKeys: []telegram.PublicKey{
			publicKey,
		},
		DC:                2,
		NoUpdates:         false,
		AllowCDN:          true,
		MigrationTimeout:  time.Second * 10,
		SessionStorage:    storage,
		UpdateHandler:     handler,
		RetryInterval:     time.Second,
		ExchangeTimeout:   time.Second * 5,
		DialTimeout:       time.Second * 5,
		EnablePFS:         true,
		TempKeyTTL:        int((time.Hour * 8).Seconds()),
		CompressThreshold: 0,
		Device: telegram.DeviceConfig{
			DeviceModel:    "DESMG Cybros",
			SystemVersion:  "macOS 26.5.2",
			AppVersion:     "0.0.1",
			SystemLangCode: "en-US",
			LangCode:       "zh",
		},
	}, nil
}

func parsePublicKey() (telegram.PublicKey, error) {
	block, _ := pem.Decode([]byte(telegramPublicKey))
	if block == nil {
		return telegram.PublicKey{}, fmt.Errorf(consts.ErrorParseTelegramPublicKeyPEMNotFound)
	}

	rsaKey, err := crypto.ParseRSA(block.Bytes)
	if err != nil {
		return telegram.PublicKey{}, fmt.Errorf(consts.ErrorParseTelegramPublicKey, err)
	}
	return telegram.PublicKey{RSA: rsaKey}, nil
}
