package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cybros/cache"
	"cybros/consts"

	"github.com/gotd/td/tg"
)

const userBioCacheTTL = 30 * time.Minute

func LoadUserBio(ctx context.Context, api *tg.Client, user tg.InputUserClass) (string, error) {
	if api == nil {
		return "", errors.New(consts.ErrorTelegramAPIUninitialized)
	}
	if user == nil {
		return "", errors.New(consts.ErrorSourceUserInputEmpty)
	}

	return cache.Load[string]("user_bio", inputUserCacheKey(user), userBioCacheTTL, func() (string, error) {
		userFull, err := RetryFloodWait(ctx, func() (*tg.UsersUserFull, error) {
			return api.UsersGetFullUser(ctx, user)
		})
		if err != nil {
			return "", err
		}
		bio, ok := userFull.FullUser.GetAbout()
		if !ok {
			return "", nil
		}
		return bio, nil
	})
}

func inputUserCacheKey(user tg.InputUserClass) string {
	inputUser, ok := user.(*tg.InputUser)
	if ok {
		return fmt.Sprintf("user:%d", inputUser.UserID)
	}

	inputUserFromMessage, ok := user.(*tg.InputUserFromMessage)
	if ok {
		return fmt.Sprintf("user:%d", inputUserFromMessage.UserID)
	}

	return fmt.Sprintf("%s:%s", user.TypeName(), user.String())
}
