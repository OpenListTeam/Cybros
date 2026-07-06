package contentaudit

import "context"

func (p *ContentAudit) handleMessage(ctx context.Context, message messageInfo) error {
	userID := message.SourceUserID
	fullNickName := message.SourceFullNickName
	userBio := message.SourceUserBio
	premiumEmojiStatusPackName := message.SourcePremiumEmojiStatusPackName
	isBot := message.SourceUserIsBot
	text := message.Text
	caption := message.Caption
	entities := message.Entities
	richTexts := message.RichTexts
	sourceGroupUsername := message.SourceGroupUsername
	sourceUserUsername := message.SourceUserUsername

	_ = userID
	_ = fullNickName
	_ = isBot
	_ = userBio
	_ = premiumEmojiStatusPackName
	_ = text
	_ = caption
	_ = entities
	_ = richTexts
	_ = sourceGroupUsername
	_ = sourceUserUsername

	return nil
}
