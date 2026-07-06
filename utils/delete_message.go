package utils

import (
	"context"
	"errors"

	"cybros/consts"

	"github.com/gotd/td/tg"
)

func DeleteMessage(ctx context.Context, api *tg.Client, channel tg.InputChannelClass, messageID int) error {
	if api == nil {
		return errors.New(consts.ErrorTelegramAPIUninitialized)
	}
	if channel == nil {
		return errors.New(consts.ErrorSourceChannelInputEmpty)
	}

	_, err := api.ChannelsDeleteMessages(ctx, &tg.ChannelsDeleteMessagesRequest{
		Channel: channel,
		ID:      []int{messageID},
	})
	return err
}
