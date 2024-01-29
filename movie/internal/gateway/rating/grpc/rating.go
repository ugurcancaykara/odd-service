package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/internal/grpcutil"
	"github.com/ugurcancaykara/odd-service/pkg/discovery"
	"github.com/ugurcancaykara/odd-service/rating/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Gateway defines an gRPC gateway for a rating service.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a rating service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

// GetAggregatedRating returns the aggregated rating for a record or ErrNotFound if there are no ratings for it.
func (g *Gateway) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	client := gen.NewRatingServiceClient(conn)
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 100 * time.Millisecond
	expBackoff.MaxInterval = 5 * time.Second
	expBackoff.MaxElapsedTime = 60 * time.Second

	// Create a retry operation with the exponential backoff strategy.
	var resp *gen.GetAggregatedRatingResponse
	operation := func() error {
		var err error
		const timeout = 10 * time.Second
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		resp, err = client.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: string(recordID), RecordType: string(recordType)})

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
		return 0, err
	}

	// If retries are exhausted and no successful response, return an error.
	if resp == nil {
		return 0, errors.New("no response received after retries")
	}

	return resp.RatingValue, nil
}

func shouldRetry(err error) bool {
	e, ok := status.FromError(err)
	if !ok {
		return false
	}
	return e.Code() == codes.DeadlineExceeded || e.Code() == codes.Unavailable || e.Code() == codes.ResourceExhausted
}

func (g *Gateway) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return nil
}
