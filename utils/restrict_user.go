package utils

import (
	"context"

	"github.com/gotd/td/tg"
)

func RestrictUser(ctx context.Context, api *tg.Client, channel tg.InputChannelClass, participant tg.InputPeerClass, untilDate int) error {
	rights := tg.ChatBannedRights{
		SendMessages:    true,
		SendMedia:       true,
		SendStickers:    true,
		SendGifs:        true,
		SendGames:       true,
		SendInline:      true,
		EmbedLinks:      true,
		SendPolls:       true,
		SendPhotos:      true,
		SendVideos:      true,
		SendRoundvideos: true,
		SendAudios:      true,
		SendVoices:      true,
		SendDocs:        true,
		SendPlain:       true,
		SendReactions:   true,
		UntilDate:       untilDate,
	}
	return editBanned(ctx, api, channel, participant, rights)
}
