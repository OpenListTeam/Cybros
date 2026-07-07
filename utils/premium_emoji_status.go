package utils

import (
	"context"
	"errors"

	"cybros/consts"

	"github.com/gotd/td/tg"
)

func EmojiStatusDocumentID(user *tg.User) int64 {
	emojiStatusClass, ok := user.GetEmojiStatus()
	if !ok {
		return 0
	}

	emojiStatus, ok := emojiStatusClass.(*tg.EmojiStatus)
	if ok {
		return emojiStatus.DocumentID
	}

	collectible, ok := emojiStatusClass.(*tg.EmojiStatusCollectible)
	if ok {
		return collectible.DocumentID
	}

	return 0
}

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

func LoadPremiumEmojiStatusPackTitleAndShortName(ctx context.Context, api *tg.Client, user *tg.User) (string, string, error) {
	if user == nil {
		return "", "", nil
	}
	if !user.GetPremium() {
		return "", "", nil
	}

	documentID := EmojiStatusDocumentID(user)
	if documentID == 0 {
		return "", "", nil
	}
	if api == nil {
		return "", "", errors.New(consts.ErrorTelegramAPIUninitialized)
	}

	return LoadCustomEmojiPackTitleAndShortName(ctx, api, documentID)
}
