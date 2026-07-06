package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"cybros/internal/config"
	"cybros/internal/logger"
	"cybros/internal/session"
	"cybros/internal/tgclient"

	"github.com/gotd/td/telegram"
)

func main() {
	logger.Init()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.Init(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		logger.Log.Fatal(err)
	}

	updateHandler := session.NewUpdateHandler()

	options, err := tgclient.NewOptions(cfg.SessionStorage, updateHandler)
	if err != nil {
		logger.Log.Fatal(err)
	}

	client := telegram.NewClient(cfg.TelegramAppID, cfg.TelegramAppHash, options)
	updateHandler.SetAPI(client.API())

	clientSession := session.New(client, cfg.Auth)

	runErr := client.Run(ctx, clientSession.Run)
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		logger.Log.Fatal(runErr)
	}
}
