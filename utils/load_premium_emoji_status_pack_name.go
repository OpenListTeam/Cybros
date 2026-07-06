package utils

import (
	"context"
	"errors"
	"strconv"
	"time"

	"cybros/cache"
	"cybros/consts"

	"github.com/gotd/td/tg"
)

const premiumEmojiStatusPackCacheTTL = 7 * 24 * time.Hour

func LoadPremiumEmojiStatusPackName(ctx context.Context, api *tg.Client, user *tg.User) (string, error) {
	if user == nil {
		return "", nil
	}
	if !user.GetPremium() {
		return "", nil
	}

	documentID := EmojiStatusDocumentID(user)
	if documentID == 0 {
		return "", nil
	}
	if api == nil {
		return "", errors.New(consts.ErrorTelegramAPIUninitialized)
	}

	return cache.Load[string]("premium_emoji_status_pack_name", strconv.FormatInt(documentID, 10), premiumEmojiStatusPackCacheTTL, func() (string, error) {
		return LoadCustomEmojiPackName(ctx, api, documentID)
	})
}
