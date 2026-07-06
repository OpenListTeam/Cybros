package utils

import "github.com/gotd/td/tg"

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
