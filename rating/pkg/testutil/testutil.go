package testutil

import (
	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/rating/internal/controller/rating"
	grpchandler "github.com/ugurcancaykara/odd-service/rating/internal/handler/grpc"
	"github.com/ugurcancaykara/odd-service/rating/internal/repository/memory"
)

// NewTestRatingGRPCServer creates a new rating gRPC server to be used in tests.
func NewTestRatingGRPCServer() gen.RatingServiceServer {
	r := memory.New()
	ctrl := rating.New(r, nil)
	return grpchandler.New(ctrl)
}
