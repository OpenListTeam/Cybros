package utils

import (
	"context"

	"github.com/gotd/td/tg"
)

func LoadPremiumEmojiStatusPack(ctx context.Context, api *tg.Client, user *tg.User) (tg.InputStickerSetClass, error) {
	if user == nil {
		return nil, nil
	}
	if !user.GetPremium() {
		return nil, nil
	}

	documentID := EmojiStatusDocumentID(user)
	if documentID == 0 {
		return nil, nil
	}

	packName, err := LoadCustomEmojiPackName(ctx, api, documentID)
	if err != nil {
		return nil, err
	}
	if packName == "" {
		return nil, nil
	}
	return &tg.InputStickerSetShortName{ShortName: packName}, nil
}
