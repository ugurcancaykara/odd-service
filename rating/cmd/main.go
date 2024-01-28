package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/pkg/discovery"
	"github.com/ugurcancaykara/odd-service/pkg/discovery/consul"
	"github.com/ugurcancaykara/odd-service/rating/internal/controller/rating"
	grpchandler "github.com/ugurcancaykara/odd-service/rating/internal/handler/grpc"
	"github.com/ugurcancaykara/odd-service/rating/internal/ingester/kafka"
	"github.com/ugurcancaykara/odd-service/rating/internal/repository/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "rating"

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port")
	flag.Parse()
	log.Printf("Starting the rating service on port %d", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	repo := memory.New()

	newIngester, err := kafka.NewIngester("localhost", "odd-service-rating-ingester", "ratings")
	if err != nil {
		fmt.Println("consumer client olustururken hata verdi")
		panic(err)
	}

	// Let's enable kafka consumer client as another goroutine, to trigger it we need to use startingestion alongside other service initialization steps
	// ctrl := rating.New(repo,nil)
	ctrl := rating.New(repo, newIngester)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterRatingServiceServer(srv, h)
	// TODO: fix this command	ctrl.StartIngestion(ctx)
	ctrl.StartIngestion(ctx)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
