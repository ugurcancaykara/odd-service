package grpc

import (
	"context"
	"errors"

	"github.com/ugurcancaykara/odd-service/gen"
	"github.com/ugurcancaykara/odd-service/metadata/internal/controller/metadata"
	"github.com/ugurcancaykara/odd-service/metadata/pkg/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a movie metadata gRPC handler.
type Handler struct {
	gen.UnimplementedMetadataServiceServer
	ctrl   *metadata.Controller
	logger *zap.Logger
}

// New creates a new movie metadata gRPC handler.
func New(ctrl *metadata.Controller, logger *zap.Logger) *Handler {
	return &Handler{ctrl: ctrl, logger: logger}
}

// GetMetadata returns movie metadata.
func (h *Handler) GetMetadata(ctx context.Context, req *gen.GetMetadataRequest) (*gen.GetMetadataResponse, error) {
	logger := h.logger.With(zap.String("method", "GetMetadata"))
	if req == nil || req.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}
	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		logger.Error("Internal error. Means some invariants expected by underlying system has been broken. You need to debug it", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &gen.GetMetadataResponse{Metadata: model.MetadataToProto(m)}, nil
}

// PutMetadata puts movie metadata to repository.
func (h *Handler) PutMetadata(ctx context.Context, req *gen.PutMetadataRequest) (*gen.PutMetadataResponse, error) {
	logger := h.logger.With(zap.String("method", "PutMetadata"))
	if req == nil || req.Metadata == nil {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or metadata")
	}
	if err := h.ctrl.Put(ctx, model.MetadataFromProto(req.Metadata)); err != nil {

		logger.Error("Internal error. Means some invariants expected by underlying system has been broken. You need to debug it", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &gen.PutMetadataResponse{}, nil
}
