package contentaudit

import (
	"context"

	"cybros/consts"
	"cybros/internal/logger"
	"cybros/utils"

	"github.com/gotd/td/tg"
)

func (p *ContentAudit) Handle(ctx context.Context, updates tg.UpdatesClass) error {
	userUsernames := map[int64]string{}
	usersByID := map[int64]*tg.User{}
	groupUsernames := map[int64]string{}
	channelPeers := map[int64]tg.InputPeerClass{}
	var users []tg.UserClass
	var chats []tg.ChatClass
	var updateList []tg.UpdateClass

	switch value := updates.(type) {
	case *tg.Updates:
		users = value.Users
		chats = value.Chats
		updateList = value.Updates
	case *tg.UpdatesCombined:
		users = value.Users
		chats = value.Chats
		updateList = value.Updates
	case *tg.UpdateShort:
		updateList = []tg.UpdateClass{value.Update}
	default:
		return nil
	}

	for _, userClass := range users {
		user, ok := userClass.(*tg.User)
		if !ok {
			continue
		}
		usersByID[user.ID] = user
		username, ok := user.GetUsername()
		if ok {
			userUsernames[user.ID] = username
		}
	}

	for _, chatClass := range chats {
		chat, ok := chatClass.(*tg.Channel)
		if !ok {
			continue
		}
		groupID := chat.ID
		username := ""
		value, ok := chat.GetUsername()
		if ok {
			username = value
		}
		groupUsernames[groupID] = username

		accessHash, ok := chat.GetAccessHash()
		if ok {
			channelPeers[groupID] = &tg.InputPeerChannel{
				ChannelID:  chat.ID,
				AccessHash: accessHash,
			}
		}
	}

	for _, update := range updateList {
		var messageClass tg.MessageClass
		channelUpdate, ok := update.(*tg.UpdateNewChannelMessage)
		if ok {
			messageClass = channelUpdate.Message
		}
		editChannelUpdate, ok := update.(*tg.UpdateEditChannelMessage)
		if ok {
			messageClass = editChannelUpdate.Message
		}
		if messageClass == nil {
			continue
		}

		message, ok := messageClass.(*tg.Message)
		if !ok {
			continue
		}

		channelPeer, ok := message.PeerID.(*tg.PeerChannel)
		if !ok {
			continue
		}
		groupID := channelPeer.ChannelID
		if groupID != 2573155438 {
			continue
		}

		userID := int64(0)
		userPeer, ok := message.FromID.(*tg.PeerUser)
		if ok {
			userID = userPeer.UserID
		}
		sourceUser := usersByID[userID]
		sourceUserInput := tg.InputUserClass(nil)
		fullNickName := ""
		isBot := false
		if sourceUser != nil {
			fullNickName = utils.FullUserName(sourceUser)
			isBot = sourceUser.GetBot()
			accessHash, ok := sourceUser.GetAccessHash()
			if ok {
				sourceUserInput = &tg.InputUser{
					UserID:     sourceUser.ID,
					AccessHash: accessHash,
				}
			}
		}
		if sourceUserInput == nil && channelPeers[groupID] != nil && userID != 0 {
			sourceUserInput = &tg.InputUserFromMessage{
				Peer:   channelPeers[groupID],
				MsgID:  message.ID,
				UserID: userID,
			}
		}

		entities, hasEntities := message.GetEntities()
		if !hasEntities {
			entities = nil
		}
		customEmojiDocumentIDs := utils.CustomEmojiDocumentIDsFromMessageEntities(entities)
		customEmojiDocumentIDs = append(customEmojiDocumentIDs, utils.CustomEmojiDocumentIDsFromMessageMedia(message.Media)...)
		text := message.Message
		caption := ""
		if message.Media != nil {
			text = ""
			caption = message.Message
		}

		auditMessage := messageInfo{
			ID:                  message.ID,
			Text:                text,
			Caption:             caption,
			SourceUserID:        userID,
			SourceFullNickName:  fullNickName,
			SourceUserBio:       "",
			SourceUserIsBot:     isBot,
			SourceGroupUsername: groupUsernames[groupID],
			SourceUserUsername:  userUsernames[userID],
		}

		if sourceUserInput != nil {
			userBio, userBioErr := utils.LoadUserBio(ctx, p.api, sourceUserInput)
			if userBioErr == nil {
				auditMessage.SourceUserBio = userBio
			} else {
				logMessageWarning(auditMessage, consts.ErrorLoadUserBio, userBioErr)
			}
		}
		premiumEmojiStatusTitle, premiumEmojiStatusPackShortName, premiumEmojiStatusPackErr := utils.LoadPremiumEmojiStatusPackTitleAndShortName(ctx, p.api, sourceUser)
		if premiumEmojiStatusPackErr == nil {
			if premiumEmojiStatusPackShortName != "" {
				auditMessage.SourcePremiumEmojiStatusTitle = premiumEmojiStatusTitle
				auditMessage.SourcePremiumEmojiStatusPackShortName = premiumEmojiStatusPackShortName
				auditMessage.SourcePremiumEmojiStatusLink = "https://t.me/addemoji/" + premiumEmojiStatusPackShortName
			}
		} else {
			logMessageWarning(auditMessage, consts.ErrorLoadPremiumEmojiStatusPackTitleAndShortName, premiumEmojiStatusPackErr)
		}
		customEmojiPackNames, customEmojiPackErr := utils.LoadCustomEmojiPackNames(ctx, p.api, customEmojiDocumentIDs)
		if customEmojiPackErr != nil {
			logMessageWarning(auditMessage, consts.ErrorLoadCustomEmojiPackNames, customEmojiPackErr)
		}
		extractedEntities := utils.ExtractMessageEntities(message.Message, entities, customEmojiPackNames)
		richTexts := utils.ExtractMessageRichTexts(message.Media, customEmojiPackNames)
		auditMessage.Entities = contentAuditMessageEntities(extractedEntities)
		auditMessage.RichTexts = contentAuditRichTexts(richTexts)

		handleErr := p.handleMessage(ctx, auditMessage)
		if handleErr != nil {
			logMessageError(auditMessage, consts.ErrorHandleContentAuditMessage, handleErr)
		}
	}

	return nil
}

func logMessageWarning(message messageInfo, text string, err error) {
	if logger.Log == nil {
		return
	}
	logger.Log.WithError(err).
		WithField("message_id", message.ID).
		WithField("source_user_id", message.SourceUserID).
		Warn(text)
}

func logMessageError(message messageInfo, text string, err error) {
	if logger.Log == nil {
		return
	}
	logger.Log.WithError(err).
		WithField("message_id", message.ID).
		WithField("source_user_id", message.SourceUserID).
		Error(text)
}
