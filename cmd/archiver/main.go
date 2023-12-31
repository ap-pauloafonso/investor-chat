package main

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/investor-chat/archive"
	"github.com/ap-pauloafonso/investor-chat/config"
	"github.com/ap-pauloafonso/investor-chat/eventbus"
	"github.com/ap-pauloafonso/investor-chat/pb"
	"github.com/ap-pauloafonso/investor-chat/storage"
	"github.com/ap-pauloafonso/investor-chat/utils"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lmittmann/tint"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"log/slog"
	"net"
	"os"
)

func main() {
	ctx := context.Background()
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	//load cfg
	var cfg config.GlobalConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		utils.LogErrorFatal(err)
	}

	// Create a listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ArchiverServerPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a gRPC server
	grpcServer := grpc.NewServer()

	// Initialize the database connection pool
	db, err := pgxpool.Connect(context.Background(), cfg.PostgresConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}

	// Close the database connection pool when the application exits
	defer db.Close()

	// create message repository
	repository := storage.NewMessageRepository(db)

	// create eventbus
	eventbus, err := eventbus.New(cfg.RabbitmqConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}
	defer eventbus.Close()
	// Create an instance of  archive service
	archiveService := archive.NewService(repository, eventbus)
	// init archiver consumer
	archiveService.InitConsumer(ctx)
	// Create an instance of gRPC service
	archiveGRPCService := NewArchiveGRPCService(archiveService)

	reflection.Register(grpcServer)

	// Register gRPC service on the gRPC server
	pb.RegisterArchiveServiceServer(grpcServer, archiveGRPCService)

	slog.Info(fmt.Sprintf("gRPC server is running on :%d", cfg.ArchiverServerPort))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

type ArchiveGRPCService struct {
	service *archive.Service
	pb.UnimplementedArchiveServiceServer
}

func NewArchiveGRPCService(service *archive.Service) *ArchiveGRPCService {
	return &ArchiveGRPCService{service: service}
}

func (s *ArchiveGRPCService) GetRecentMessages(ctx context.Context, req *pb.GetRecentMessagesRequest) (*pb.GetRecentMessagesResponse, error) {
	messages, err := s.service.GetRecentMessages(ctx, req.Channel)
	if err != nil {
		return nil, err
	}

	r := make([]*pb.Message, len(messages))

	for i := range messages {

		var item = messages[i]
		r[i] = &pb.Message{
			Channel:   item.Channel,
			User:      item.User,
			Text:      item.Text,
			Timestamp: timestamppb.New(item.Timestamp),
		}

	}
	return &pb.GetRecentMessagesResponse{Messages: r}, nil
}
