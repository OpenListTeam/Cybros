package utils

import (
	"context"
	"errors"

	"cybros/consts"

	"github.com/gotd/td/tg"
)

func editBanned(ctx context.Context, api *tg.Client, channel tg.InputChannelClass, participant tg.InputPeerClass, rights tg.ChatBannedRights) error {
	if api == nil {
		return errors.New(consts.ErrorTelegramAPIUninitialized)
	}
	if channel == nil {
		return errors.New(consts.ErrorSourceChannelInputEmpty)
	}
	if participant == nil {
		return errors.New(consts.ErrorSourceParticipantEmpty)
	}

	_, err := RetryFloodWait(ctx, func() (struct{}, error) {
		_, err := api.ChannelsEditBanned(ctx, &tg.ChannelsEditBannedRequest{
			Channel:      channel,
			Participant:  participant,
			BannedRights: rights,
		})
		return struct{}{}, err
	})
	return err
}
