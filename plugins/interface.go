package plugins

import (
	"context"

	"github.com/gotd/td/tg"
)

type Plugin interface {
	Handle(ctx context.Context, updates tg.UpdatesClass) error
}

type APIPlugin interface {
	SetAPI(api *tg.Client)
}
