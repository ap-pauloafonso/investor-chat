package main

import (
	"github.com/lmittmann/tint"
	"investorchat/bot"
	"investorchat/queue"
	"investorchat/utils"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	slog.Info("starting the bot")

	q, err := queue.NewQueue()
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
