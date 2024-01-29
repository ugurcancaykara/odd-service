package rating

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ugurcancaykara/odd-service/rating/internal/repository"
	model "github.com/ugurcancaykara/odd-service/rating/pkg/model"
)

// ErrNotFound is returned when no ratings are found for a record.
var ErrNotFound = errors.New("ratings not found for a record")

type ratingRepository interface {
	Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error
}

type ratingIngester interface {
	Ingest(ctx context.Context) (chan model.RatingEvent, error)
}

// Controller defines a rating service controller.
type Controller struct {
	repo     ratingRepository
	ingester ratingIngester
}

// New creates a rating service controller.
func New(repo ratingRepository, ingester ratingIngester) *Controller {
	return &Controller{repo, ingester}
}

// GetAggregatedRating returns the aggregated rating for a record or ErrNotFound if there are no ratings for it.
func (c *Controller) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	ratings, err := c.repo.Get(ctx, recordID, recordType)
	if err != nil && err == repository.ErrNotFound {
		return 0, ErrNotFound
	} else if err != nil {
		log.Printf("Failed to get ratings for %v %v: %v", recordID, recordType, err)
		log.Printf("Fallback: returning locally cached ratings for %v %v", recordID, recordType)
		// Fallback scenario to increase reliability of a service,
		// return locally cached ratings in case of mysql db is unavailable
		// Using fallbacks is an example of graceful degradation - a pracice of handling application failures in a wa
		// that an application still performs its operations in a limited mode. the movie service would continue processing
		// requests for getting movie details even if the recommendation feature is unavailable, providing a limited but working
		// functionality to its users

		// TODO: implement in-memory cache solution
		// return c.getCachedRatings(recordID, recordType)
		return 0, err
	}
	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}
	return sum / float64(len(ratings)), nil
}

// PutRating writes a rating for a given record.
func (c *Controller) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return c.repo.Put(ctx, recordID, recordType, rating)
}

// StartIngestion starts the ingestion of rating events.
func (s *Controller) StartIngestion(ctx context.Context) error {
	ch, err := s.ingester.Ingest(ctx)
	if err != nil {
		return err
	}
	fmt.Println("starting to ingestion")
	for e := range ch {
		if err := s.PutRating(ctx, e.RecordID, e.RecordType, &model.Rating{UserID: e.UserID, Value: e.Value}); err != nil {
			return err
		}
	}
	return nil
}

// At this point, the rating service provides both a synchronous API for the callers that
// want to create ratings in real time and asynchronous logic for ingesting
// rating events from Apache Kafka.
