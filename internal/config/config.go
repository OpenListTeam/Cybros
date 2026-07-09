package config

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cybros/consts"
	"cybros/internal/console"
	"cybros/internal/sessionstore"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
)

const (
	defaultTelegramAPIID = 8213735

	envTelegramAPIID   = "TELEGRAM_API_ID"
	envTelegramAPIHash = "TELEGRAM_API_HASH"
	envSessionPassword = "CYBROS_SESSION_PASSWORD"
)

type Config struct {
	TelegramAppID   int
	TelegramAppHash string
	Auth            auth.UserAuthenticator
	SessionStorage  telegram.SessionStorage
}

func Init(ctx context.Context) (Config, error) {
	reader := bufio.NewReader(os.Stdin)

	telegramAppID, err := telegramAPIID()
	if err != nil {
		return Config{}, err
	}

	telegramAppHash, err := telegramAPIHash(ctx, reader)
	if err != nil {
		return Config{}, err
	}

	sessionPassword, err := sessionPassword(ctx, reader)
	if err != nil {
		return Config{}, err
	}

	storage, err := sessionstore.Init(sessionPassword)
	if err != nil {
		return Config{}, err
	}

	return Config{
		TelegramAppID:   telegramAppID,
		TelegramAppHash: telegramAppHash,
		Auth:            console.NewAuth(reader),
		SessionStorage:  storage,
	}, nil
}

func telegramAPIID() (int, error) {
	value := strings.TrimSpace(os.Getenv(envTelegramAPIID))
	if value == "" {
		return defaultTelegramAPIID, nil
	}
	return validateTelegramAPIID(value)
}

func telegramAPIHash(ctx context.Context, reader *bufio.Reader) (string, error) {
	value := strings.TrimSpace(os.Getenv(envTelegramAPIHash))
	if value == "" {
		return console.ReadTelegramAppHash(ctx, reader)
	}
	validateErr := console.ValidateTelegramAppHash(value)
	if validateErr != nil {
		return "", validateErr
	}
	return value, nil
}

func sessionPassword(ctx context.Context, reader *bufio.Reader) (string, error) {
	value := strings.TrimSpace(os.Getenv(envSessionPassword))
	if value != "" {
		return value, nil
	}
	return console.ReadSessionPassphrase(ctx, reader)
}

func validateTelegramAPIID(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(consts.ErrorTelegramAPIIDInteger, err)
	}
	if id <= 0 {
		return 0, errors.New(consts.ErrorTelegramAPIIDPositive)
	}
	return id, nil
}
