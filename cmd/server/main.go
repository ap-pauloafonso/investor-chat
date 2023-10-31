package main

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/investor-chat/channel"
	"github.com/ap-pauloafonso/investor-chat/config"
	"github.com/ap-pauloafonso/investor-chat/eventbus"
	"github.com/ap-pauloafonso/investor-chat/frontend"
	"github.com/ap-pauloafonso/investor-chat/pb"
	"github.com/ap-pauloafonso/investor-chat/server"
	"github.com/ap-pauloafonso/investor-chat/storage"
	"github.com/ap-pauloafonso/investor-chat/user"
	"github.com/ap-pauloafonso/investor-chat/utils"
	"github.com/ap-pauloafonso/investor-chat/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	// load cfg
	var cfg config.GlobalConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		utils.LogErrorFatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("starting the server...")

	// Initialize the database connection pool
	db, err := pgxpool.Connect(context.Background(), cfg.PostgresConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}

	// Close the database connection pool when the application exits
	defer db.Close()

	// create event bus
	eventbus, err := eventbus.New(cfg.RabbitmqConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}
	defer eventbus.Close()

	// create user repository
	userRepository := storage.NewUserRepository(db)
	// Create the user service
	userService := user.NewService(userRepository)

	// create channel repository
	channelRepository := storage.NewChannelRepository(db)

	grpcConn, err := grpc.Dial(cfg.GrpcConnection, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.LogErrorFatal(fmt.Errorf("error connecting to grpc server: %w", err))
	}
	defer grpcConn.Close()

	grpcClient := pb.NewArchiveServiceClient(grpcConn)

	// create websocket handler
	wserver := websocket.NewWebSocketHandler(eventbus, grpcClient)
	// start printing the sessions
	wserver.PrintOnlineUsers()

	// create channel service
	channelService := channel.NewService(channelRepository, eventbus, wserver)

	// Create the application instance
	server := server.NewApp(ctx, userService, channelService, eventbus, frontend.FS, wserver)

	// Start the server
	go func() {
		slog.Info(fmt.Sprintf("server is running on :%d", cfg.ServerPort))
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
