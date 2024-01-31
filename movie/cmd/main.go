package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/movie/internal/controller/movie"
	metadatagateway "github.com/ugurcancaykara/odd-service/movie/internal/gateway/metadata/grpc"
	ratinggateway "github.com/ugurcancaykara/odd-service/movie/internal/gateway/rating/grpc"
	grpchandler "github.com/ugurcancaykara/odd-service/movie/internal/handler/grpc"
	"github.com/ugurcancaykara/odd-service/pkg/discovery"
	"github.com/ugurcancaykara/odd-service/pkg/discovery/consul"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "movie"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/base.yaml"
	}
	f, err := os.Open(configPath)
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))

	}
	defer f.Close()
	var cfg serviceConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))

	}
	port := cfg.API.Port

	logger.Info("Starting the movie service", zap.Int("port", port), zap.String("serviceName", serviceName))

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))

			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	h := grpchandler.New(ctrl, logger)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterMovieServiceServer(srv, h)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		log.Printf("Received signal %v, attempting graceful shutdown...", s)
		srv.GracefulStop()
		log.Println("Gracefully stopped gRPC server")
	}()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
