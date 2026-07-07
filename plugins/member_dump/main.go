package memberdump

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cybros/internal/logger"
	"cybros/utils"

	"github.com/gotd/td/tg"
)

const (
	command       = ".cl"
	targetGroupID = int64(2573155438)

	pageLimit    = 100
	requestDelay = time.Second
)

type dumpTrigger struct {
	channel tg.InputChannelClass
	msgID   int
}

func (p *MemberDump) Handle(ctx context.Context, updates tg.UpdatesClass) error {
	var chats []tg.ChatClass
	var updateList []tg.UpdateClass

	switch value := updates.(type) {
	case *tg.Updates:
		chats = value.Chats
		updateList = value.Updates
	case *tg.UpdatesCombined:
		chats = value.Chats
		updateList = value.Updates
	case *tg.UpdateShort:
		updateList = []tg.UpdateClass{value.Update}
	default:
		return nil
	}

	channels := inputChannels(chats)
	for _, update := range updateList {
		trigger, ok := dumpTriggerFromUpdate(update, channels)
		if !ok {
			continue
		}
		p.startDump(ctx, trigger)
	}

	return nil
}

func inputChannels(chats []tg.ChatClass) map[int64]tg.InputChannelClass {
	channels := map[int64]tg.InputChannelClass{}
	for _, chatClass := range chats {
		chat, ok := chatClass.(*tg.Channel)
		if !ok {
			continue
		}

		accessHash, ok := chat.GetAccessHash()
		if !ok {
			continue
		}

		channels[chat.ID] = &tg.InputChannel{
			ChannelID:  chat.ID,
			AccessHash: accessHash,
		}
	}

	return channels
}

func dumpTriggerFromUpdate(update tg.UpdateClass, channels map[int64]tg.InputChannelClass) (dumpTrigger, bool) {
	channelUpdate, ok := update.(*tg.UpdateNewChannelMessage)
	if !ok {
		return dumpTrigger{}, false
	}

	message, ok := channelUpdate.Message.(*tg.Message)
	if !ok {
		return dumpTrigger{}, false
	}
	if !message.Out {
		return dumpTrigger{}, false
	}
	if strings.TrimSpace(message.Message) != command {
		return dumpTrigger{}, false
	}

	channelPeer, ok := message.PeerID.(*tg.PeerChannel)
	if !ok {
		return dumpTrigger{}, false
	}
	if channelPeer.ChannelID != targetGroupID {
		return dumpTrigger{}, false
	}

	channel := channels[channelPeer.ChannelID]
	if channel == nil {
		logWarn(message.ID, "Member dump channel input is empty")
		return dumpTrigger{}, false
	}

	return dumpTrigger{
		channel: channel,
		msgID:   message.ID,
	}, true
}

func (p *MemberDump) startDump(ctx context.Context, trigger dumpTrigger) {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		logWarn(trigger.msgID, "Member dump is already running")
		return
	}
	p.running = true
	p.mu.Unlock()

	defer p.finishDump()

	err := p.dumpMembers(ctx, trigger)
	if err != nil {
		logError(trigger.msgID, err)
	}
}

func (p *MemberDump) finishDump() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.running = false
}

func (p *MemberDump) dumpMembers(ctx context.Context, trigger dumpTrigger) error {
	if p.api == nil {
		return fmt.Errorf("Telegram API is not initialized")
	}

	path, err := newDumpPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("Open member dump file %s: %w", path, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	if logger.Log != nil {
		logger.Log.WithField("path", path).Info("member dump started")
	}

	offset := 0
	written := 0
	for {
		result, err := utils.RetryFloodWait(ctx, func() (*tg.ChannelsChannelParticipants, error) {
			return p.loadParticipants(ctx, trigger.channel, offset)
		})
		if err != nil {
			return fmt.Errorf("Get channel participants offset %d: %w", offset, err)
		}
		if result == nil {
			break
		}
		if len(result.Participants) == 0 {
			break
		}

		users := usersByID(result.Users)
		premiumEmojiStatusTitles, premiumEmojiStatusPackShortNames, err := p.loadPremiumEmojiStatusPacks(ctx, trigger, result.Users)
		if err != nil {
			return err
		}
		for _, participant := range result.Participants {
			participantUserID := userID(participant)
			user := users[participantUserID]
			err = writeMemberLine(writer, participant, user, premiumEmojiStatusTitles[participantUserID], premiumEmojiStatusPackShortNames[participantUserID])
			if err != nil {
				return fmt.Errorf("Write member dump file %s: %w", path, err)
			}
			written++
		}

		offset += len(result.Participants)
		if len(result.Participants) < pageLimit {
			break
		}

		err = sleepContext(ctx, requestDelay)
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("Flush member dump file %s: %w", path, err)
	}

	if logger.Log != nil {
		logger.Log.WithField("path", path).
			WithField("members", written).
			Info("member dump finished")
	}

	return nil
}

func (p *MemberDump) loadParticipants(ctx context.Context, channel tg.InputChannelClass, offset int) (*tg.ChannelsChannelParticipants, error) {
	result, err := p.api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
		Channel: channel,
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  offset,
		Limit:   pageLimit,
		Hash:    0,
	})
	if err != nil {
		return nil, err
	}

	participants, ok := result.(*tg.ChannelsChannelParticipants)
	if !ok {
		return nil, nil
	}

	return participants, nil
}

func (p *MemberDump) loadPremiumEmojiStatusPacks(ctx context.Context, trigger dumpTrigger, users []tg.UserClass) (map[int64]string, map[int64]string, error) {
	packTitles, packShortNames, err := p.premiumEmojiStatusPacks(ctx, users)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, err
		}
		logWarningWithError(trigger.msgID, "Load premium emoji status failed", err)
		return nil, nil, nil
	}

	return packTitles, packShortNames, nil
}

func (p *MemberDump) premiumEmojiStatusPacks(ctx context.Context, users []tg.UserClass) (map[int64]string, map[int64]string, error) {
	userDocumentIDs := map[int64]int64{}
	documentIDs := []int64{}

	for _, userClass := range users {
		user, ok := userClass.(*tg.User)
		if !ok {
			continue
		}

		documentID := utils.EmojiStatusDocumentID(user)
		if documentID == 0 {
			continue
		}

		userDocumentIDs[user.ID] = documentID
		documentIDs = append(documentIDs, documentID)
	}
	if len(documentIDs) == 0 {
		return nil, nil, nil
	}

	packTitles, packShortNames, err := utils.LoadCustomEmojiPackTitlesAndShortNames(ctx, p.api, documentIDs)
	if err != nil {
		return nil, nil, err
	}

	packTitlesByUserID := map[int64]string{}
	packShortNamesByUserID := map[int64]string{}
	for userID, documentID := range userDocumentIDs {
		packShortName := packShortNames[documentID]
		if packShortName == "" {
			continue
		}

		packTitlesByUserID[userID] = packTitles[documentID]
		packShortNamesByUserID[userID] = packShortName
	}

	return packTitlesByUserID, packShortNamesByUserID, nil
}

func usersByID(users []tg.UserClass) map[int64]tg.UserClass {
	result := map[int64]tg.UserClass{}
	for _, user := range users {
		if user == nil {
			continue
		}
		result[user.GetID()] = user
	}
	return result
}

func userID(participant tg.ChannelParticipantClass) int64 {
	if participant == nil {
		return 0
	}

	value := reflect.ValueOf(participant)
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return 0
		}
		value = value.Elem()
	}

	field := value.FieldByName("UserID")
	if !field.IsValid() {
		return 0
	}
	if field.Kind() != reflect.Int64 {
		return 0
	}
	return field.Int()
}

func writeMemberLine(writer *bufio.Writer, participant tg.ChannelParticipantClass, user tg.UserClass, premiumEmojiStatusTitle string, premiumEmojiStatusPackShortName string) error {
	line := map[string]any{}

	participantObject, ok := participant.(tlObject)
	if ok {
		line["participant"] = marshalTLObject(participantObject)
	}

	userObject, ok := user.(tlObject)
	if ok {
		line["user"] = marshalTLObject(userObject)
	}
	if premiumEmojiStatusPackShortName != "" {
		line["premium_emoji_status_title"] = premiumEmojiStatusTitle
		line["premium_emoji_status_pack_short_name"] = premiumEmojiStatusPackShortName
		line["premium_emoji_status_link"] = "https://t.me/addemoji/" + premiumEmojiStatusPackShortName
	}

	data, err := json.Marshal(line)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	err = writer.WriteByte('\n')
	if err != nil {
		return err
	}

	return nil
}

func newDumpPath() (string, error) {
	dir := filepath.Join("data", "member_dump")
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "", fmt.Errorf("Create member dump dir %s: %w", dir, err)
	}

	fileName := "channel_" + strconv.FormatInt(targetGroupID, 10) + "_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".jsonl"
	return filepath.Join(dir, fileName), nil
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func logWarn(messageID int, text string) {
	if logger.Log == nil {
		return
	}
	logger.Log.WithField("message_id", messageID).Warn(text)
}

func logWarningWithError(messageID int, text string, err error) {
	if logger.Log == nil {
		return
	}
	logger.Log.WithError(err).WithField("message_id", messageID).Warn(text)
}

func logError(messageID int, err error) {
	if logger.Log == nil {
		return
	}
	logger.Log.WithError(err).WithField("message_id", messageID).Error("Member dump failed")
}
