package tgclient

import (
	"context"

	"github.com/gotd/td/tg"
)

type updateHandler struct{}

func (h updateHandler) Handle(ctx context.Context, update tg.UpdatesClass) error {
	return nil
}
