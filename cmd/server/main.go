package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"investorchat/app"
	"investorchat/chat"
	assets "investorchat/frontend"
	"investorchat/queue"
	"investorchat/storage"
	"investorchat/user"
	"log"
)

func main() {
	// Initialize the database connection pool
	db, err := pgxpool.Connect(context.Background(), getDatabaseDSN())
	if err != nil {
		log.Fatal(err)
	}

	// Close the database connection pool when the application exits
	defer db.Close()

	// Create the repository for user storage
	userRepository := storage.NewUserRepository(db)

	// Create the user service
	userService := user.NewService(userRepository)

	// Create the repository for chat storage
	chatRepository := storage.NewChatRepository(db)

	// Create the chat service
	chatService := chat.NewService(chatRepository)

	queue, err := queue.NewQueue()
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	// Create the application instance
	chatApp := app.NewApp(&userService, &chatService, queue, assets.FrontendFS)

	// Start the server
	port := "8080"
	err = chatApp.E.Start(":" + port)
	if err != nil {
		log.Fatal(err)
	}
}

func getDatabaseDSN() string {
	return "host=postgres port=5432 user=postgres password=test dbname=MY_DB sslmode=disable"
}
