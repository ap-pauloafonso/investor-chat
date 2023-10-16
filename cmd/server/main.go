package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	"investorchat/chat"
	"investorchat/frontend"
	"investorchat/queue"
	"investorchat/server"
	"investorchat/storage"
	"investorchat/user"
	"investorchat/utils"
	"investorchat/websocket"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	dbConnection = "host=postgres port=5432 user=postgres password=test dbname=MY_DB sslmode=disable"
	serverPort   = 8080
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	slog.Info("starting the server")

	// Initialize the database connection pool
	db, err := pgxpool.Connect(context.Background(), dbConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}

	// Close the database connection pool when the application exits
	defer db.Close()

	// Create the repository for user storage
	userRepository := storage.NewUserRepository(db)

	// Create the user service
	userService := user.NewService(userRepository)

	// Create the repository for chat storage
	chatRepository := storage.NewChatRepository(db)

	queue, err := queue.NewQueue()
	if err != nil {
		utils.LogErrorFatal(err)
	}
	defer queue.Close()

	// create socker handler server
	wserver := websocket.NewWebSocketHandler(queue)
	chatService := chat.NewService(chatRepository, queue, wserver)
	wserver.OnConnection(chatService.UserConnected)
	wserver.PrintOnlineUsers()
	// Create the application instance
	server := server.NewApp(&userService, &chatService, queue, frontend.FS, wserver)

	// Start the server

	go func() {
		if err := server.E.Start(fmt.Sprintf(":%d", serverPort)); err != nil {
			utils.LogErrorFatal(err)
		}
	}()

	// Wait for a signal to exit
	sig := <-c

	// Shutdown the server gracefully
	if err := server.E.Shutdown(context.Background()); err != nil {
		utils.LogErrorFatal(err)
	}

	slog.Info("Received signal, Server shut down gracefully", sig)

}
