package testutil

import (
	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/movie/internal/controller/movie"
	metadatagateway "github.com/ugurcancaykara/odd-service/movie/internal/gateway/metadata/grpc"
	ratinggateway "github.com/ugurcancaykara/odd-service/movie/internal/gateway/rating/grpc"
	grpchandler "github.com/ugurcancaykara/odd-service/movie/internal/handler/grpc"
	"github.com/ugurcancaykara/odd-service/pkg/discovery"
)

// NewTestMovieGRPCServer creates a new movie gRPC server to be used in tests.
func NewTestMovieGRPCServer(registry discovery.Registry) gen.MovieServiceServer {
	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	return grpchandler.New(ctrl)
}
