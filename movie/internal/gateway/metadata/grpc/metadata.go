package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/internal/grpcutil"
	"github.com/ugurcancaykara/odd-service/metadata/pkg/model"
	"github.com/ugurcancaykara/odd-service/pkg/discovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Gateway defines a movie metadata gRPC gateway.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a movie metadata service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func shouldRetry(err error) bool {
	e, ok := status.FromError(err)
	if !ok {
		return false
	}
	return e.Code() == codes.DeadlineExceeded || e.Code() == codes.Unavailable || e.Code() == codes.ResourceExhausted
}

// Get returns movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)

	// Create an exponential backoff with jittering strategy.
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 100 * time.Millisecond // Initial retry interval
	expBackoff.MaxInterval = 5 * time.Second            // Maximum retry interval
	expBackoff.MaxElapsedTime = 60 * time.Second        // Maximum total time for retries
	expBackoff.RandomizationFactor = 0.5                // jittering for randomization

	// Create a retry operation with the exponential backoff strategy.
	var resp *gen.GetMetadataResponse
	operation := func() error {
		var err error
		const timeout = 10 * time.Second
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		resp, err = client.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: id})
		if err != nil {
			if shouldRetry(err) {
				return err // Retry
			}
			return backoff.Permanent(err) // Stop retrying on non-retriable error
		}
		return nil // Success, stop retrying
	}

	// Use the backoff.Retry function to perform the retry logic.
	if err := backoff.Retry(operation, expBackoff); err != nil {
		return nil, err
	}

	// If retries are exhausted and no successful response, return an error.
	if resp == nil {
		return nil, errors.New("no response received after retries")
	}

	// Return the metadata extracted from the response.
	return model.MetadataFromProto(resp.Metadata), nil
}
