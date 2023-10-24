package main

import (
	"context"
	"github.com/lmittmann/tint"
	"github.com/sethvargo/go-envconfig"
	"investorchat/bot"
	"investorchat/config"
	"investorchat/queue"
	"investorchat/utils"
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

	q, err := queue.NewQueue(cfg.RabbitmqConnection)
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
