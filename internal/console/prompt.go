package console

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"cybros/consts"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type Auth struct {
	reader *bufio.Reader
}

func NewAuth(reader *bufio.Reader) *Auth {
	return &Auth{
		reader: reader,
	}
}

func (a *Auth) Phone(ctx context.Context) (string, error) {
	return askRequired(ctx, a.reader, "Phone: ", consts.ErrorEmptyPhone, false)
}

func (a *Auth) Password(ctx context.Context) (string, error) {
	return askRequired(ctx, a.reader, "2FA password: ", consts.ErrorEmpty2FAPassword, true)
}

func (a *Auth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	return askRequired(ctx, a.reader, "Login code: ", consts.ErrorEmptyLoginCode, false)
}

func (a *Auth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return errors.New(consts.ErrorSignUpNotSupported)
}

func (a *Auth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New(consts.ErrorSignUpNotSupported)
}

func ReadSessionPassphrase(ctx context.Context, reader *bufio.Reader) (string, error) {
	return askRequired(ctx, reader, "Session password: ", consts.ErrorEmptySessionPassword, true)
}

func ReadTelegramAppHash(ctx context.Context, reader *bufio.Reader) (string, error) {
	hash, err := askRequired(ctx, reader, "Telegram API hash: ", consts.ErrorEmptyTelegramAPIHash, true)
	if err != nil {
		return "", err
	}
	validateErr := ValidateTelegramAppHash(hash)
	if validateErr != nil {
		return "", validateErr
	}
	return hash, nil
}

func ValidateTelegramAppHash(hash string) error {
	if len(hash) != 32 {
		return fmt.Errorf(consts.ErrorTelegramAPIHashLength, len(hash))
	}
	for _, c := range hash {
		if !isHex(c) {
			return errors.New(consts.ErrorTelegramAPIHashMustBeHex)
		}
	}
	return nil
}

type readResult struct {
	text string
	err  error
}

func askRequired(ctx context.Context, reader *bufio.Reader, prompt, emptyMessage string, hidden bool) (string, error) {
	fmt.Print(prompt)

	var (
		text string
		err  error
	)
	if hidden {
		text, err = readHiddenLine(ctx, reader)
	} else {
		text, err = readLine(ctx, reader)
	}
	if err != nil {
		return "", err
	}

	value := strings.TrimSpace(text)
	if value == "" {
		return "", errors.New(emptyMessage)
	}
	return value, nil
}

func readLine(ctx context.Context, reader *bufio.Reader) (string, error) {
	done := make(chan readResult, 1)
	go func() {
		text, err := reader.ReadString('\n')
		done <- readResult{
			text: text,
			err:  err,
		}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-done:
		return result.text, result.err
	}
}

func isHex(c rune) bool {
	return c >= '0' && c <= '9' ||
		c >= 'a' && c <= 'f' ||
		c >= 'A' && c <= 'F'
}
