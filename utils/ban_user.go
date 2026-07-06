package utils

import (
	"context"

	"github.com/gotd/td/tg"
)

func BanUser(ctx context.Context, api *tg.Client, channel tg.InputChannelClass, participant tg.InputPeerClass) error {
	rights := tg.ChatBannedRights{
		ViewMessages: true,
	}
	return editBanned(ctx, api, channel, participant, rights)
}
