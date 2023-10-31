package main

import (
	"context"
	"github.com/ap-pauloafonso/investor-chat/bot"
	"github.com/ap-pauloafonso/investor-chat/config"
	"github.com/ap-pauloafonso/investor-chat/eventbus"
	"github.com/ap-pauloafonso/investor-chat/utils"
	"github.com/lmittmann/tint"
	"github.com/sethvargo/go-envconfig"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))
	ctx := context.Background()
	var cfg config.GlobalConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		utils.LogErrorFatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("starting the bot")

	q, err := eventbus.New(cfg.RabbitmqConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}
	defer q.Close()

	// start processing
	err = bot.NewBot(q).Process()
	if err != nil {
		utils.LogErrorFatal(err)
	}

	// Block until a signal is received
	sig := <-c

	slog.Info("Received signal, Server shut down gracefully", sig)

}
