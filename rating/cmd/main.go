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
	"github.com/ugurcancaykara/odd-service/pkg/discovery"
	"github.com/ugurcancaykara/odd-service/pkg/discovery/consul"
	"github.com/ugurcancaykara/odd-service/pkg/tracing"
	"github.com/ugurcancaykara/odd-service/rating/internal/controller/rating"
	grpchandler "github.com/ugurcancaykara/odd-service/rating/internal/handler/grpc"
	"github.com/ugurcancaykara/odd-service/rating/internal/ingester/kafka"
	"github.com/ugurcancaykara/odd-service/rating/internal/repository/mysql"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "rating"

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
	if err = yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))

	}
	port := cfg.API.Port
	logger.Info("Starting the movie service", zap.Int("port", port), zap.String("serviceName", serviceName))

	ctx, cancel := context.WithCancel(context.Background())
	tp, err := tracing.NewJaegerProvider(cfg.Jaeger.URL, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize Jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Fatal("Failed to shut down Jaeger provider", zap.Error(err))
		}
	}()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
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
	repo, err := mysql.New()
	if err != nil {
		panic(err)
	}
	newIngester, err := kafka.NewIngester("localhost", "odd-service-rating-ingester", "ratings")
	if err != nil {
		logger.Fatal("Failed to create kafka ingester", zap.Error(err))
	}
	ctrl := rating.New(repo, newIngester)
	go ctrl.StartIngestion(ctx)

	h := grpchandler.New(ctrl, logger)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	reflection.Register(srv)
	gen.RegisterRatingServiceServer(srv, h)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("Failed to shut down Jaeger provider", zap.Error(err))
		}
		cancel()
		log.Printf("Received signal %v, attempting graceful shutdown...", s)
		srv.GracefulStop()
		log.Println("Gracefully stopped gRPC server")
	}()
	fmt.Println("starting grpc server ")
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	wg.Wait()
}
