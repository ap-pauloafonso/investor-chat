package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"investorchat/channel"
	"investorchat/config"
	"investorchat/frontend"
	"investorchat/pb"
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

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	ctx := context.Background()
	var cfg config.GlobalConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		utils.LogErrorFatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("starting the server")

	// Initialize the database connection pool
	db, err := pgxpool.Connect(context.Background(), cfg.PostgresConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}

	// Close the database connection pool when the application exits
	defer db.Close()

	queue, err := queue.NewQueue(cfg.RabbitmqConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}
	defer queue.Close()

	userRepository := storage.NewUserRepository(db)
	// Create the user service
	userService := user.NewService(userRepository)

	channelRepository := storage.NewChannelRepository(db)

	grpcConn, err := grpc.Dial(cfg.GrpcConnection, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.LogErrorFatal(fmt.Errorf("error connecting to grpc server: %w", err))
	}
	defer grpcConn.Close()

	grpcClient := pb.NewArchiveServiceClient(grpcConn)

	wserver := websocket.NewWebSocketHandler(queue, grpcClient)
	wserver.PrintOnlineUsers()
	channelService := channel.NewService(channelRepository, queue, wserver)

	// Create the application instance
	server := server.NewApp(userService, channelService, queue, frontend.FS, wserver)

	// Start the server

	go func() {
		if err := server.E.Start(fmt.Sprintf(":%d", cfg.ServerPort)); err != nil {
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
