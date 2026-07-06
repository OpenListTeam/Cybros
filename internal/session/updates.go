package session

import (
	"context"

	"cybros/consts"
	"cybros/internal/logger"
	"cybros/plugins"

	"github.com/gotd/td/tg"
)

type UpdateHandler struct {
	plugins []plugins.Plugin
}

func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{
		plugins: plugins.New(),
	}
}

func (h *UpdateHandler) SetAPI(api *tg.Client) {
	plugins.SetAPI(h.plugins, api)
}

func (h *UpdateHandler) Handle(ctx context.Context, updates tg.UpdatesClass) error {
	for _, plugin := range h.plugins {
		err := plugin.Handle(ctx, updates)
		if err != nil {
			logger.Log.WithError(err).Error(consts.ErrorTelegramPlugin)
		}
	}

	return nil
}
