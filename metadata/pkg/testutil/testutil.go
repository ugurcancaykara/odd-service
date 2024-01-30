package testutil

import (
	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/metadata/internal/controller/metadata"
	grpchandler "github.com/ugurcancaykara/odd-service/metadata/internal/handler/grpc"
	"github.com/ugurcancaykara/odd-service/metadata/internal/repository/memory"
)

// NewTestMetadataGRPCServer creates a new metadata gRPC server to be used in tests.
func NewTestMetadataGRPCServer() gen.MetadataServiceServer {
	r := memory.New()
	ctrl := metadata.New(r)
	return grpchandler.New(ctrl)
}
