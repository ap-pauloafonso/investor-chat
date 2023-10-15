package main

import (
	"fmt"
	"investorchat/bot"
	"investorchat/queue"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	q, err := queue.NewQueue()
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	// start processing
	bot.NewBot(q).Process()

	// Block until a signal is received
	sig := <-c
	fmt.Printf("Received signal: %v\n", sig)

	// Exit the program
	os.Exit(0)

}
