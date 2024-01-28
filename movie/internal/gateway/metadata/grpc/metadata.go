package grpc

import (
	"context"

	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/internal/grpcutil"
	"github.com/ugurcancaykara/odd-service/metadata/pkg/model"
	"github.com/ugurcancaykara/odd-service/pkg/discovery"
)

// Gateway defines a movie metadata gRPC gateway.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a movie metadata service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

// Get returns movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: id})
	if err != nil {
		return nil, err
	}
	return model.MetadataFromProto(resp.Metadata), nil
}
