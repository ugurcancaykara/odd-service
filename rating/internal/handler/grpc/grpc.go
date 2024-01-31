package grpc

import (
	"context"
	"errors"

	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/rating/internal/controller/rating"
	"github.com/ugurcancaykara/odd-service/rating/pkg/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a gRPC rating API handler.
type Handler struct {
	gen.UnimplementedRatingServiceServer
	ctrl   *rating.Controller
	logger *zap.Logger
}

// New creates a new movie metadata gRPC handler.
func New(ctrl *rating.Controller, logger *zap.Logger) *Handler {
	return &Handler{ctrl: ctrl, logger: logger}
}

// GetAggregatedRating returns the aggregated rating for a record.
func (h *Handler) GetAggregatedRating(ctx context.Context, req *gen.GetAggregatedRatingRequest) (*gen.GetAggregatedRatingResponse, error) {
	logger := h.logger.With(zap.String("method", "GetAggregatedRating"))
	if req == nil || req.RecordId == "" || req.RecordType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}
	v, err := h.ctrl.GetAggregatedRating(ctx, model.RecordID(req.RecordId), model.RecordType(req.RecordType))
	if err != nil && errors.Is(err, rating.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		logger.Error("Internal error. Means some invariants expected by underlying system has been broken. You need to debug it", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &gen.GetAggregatedRatingResponse{RatingValue: v}, nil
}

// PutRating writes a rating for a given record.
func (h *Handler) PutRating(ctx context.Context, req *gen.PutRatingRequest) (*gen.PutRatingResponse, error) {

	logger := h.logger.With(zap.String("method", "PutRating"))
	if req == nil || req.RecordId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty user id or record id")
	}
	if err := h.ctrl.PutRating(ctx, model.RecordID(req.RecordId), model.RecordType(req.RecordType), &model.Rating{UserID: model.UserID(req.UserId), Value: model.RatingValue(req.RatingValue)}); err != nil {
		logger.Error("Internal error. Means some invariants expected by underlying system has been broken. You need to debug it", zap.Error(err))
		return nil, err
	}
	return &gen.PutRatingResponse{}, nil
}
