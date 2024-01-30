package rating

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	gen "github.com/ugurcancaykara/odd-service/gen/mock/rating/repository"
	"github.com/ugurcancaykara/odd-service/rating/internal/repository"
	model "github.com/ugurcancaykara/odd-service/rating/pkg/model"
)

func TestController_GetAggregatedRating(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := gen.NewMockratingRepository(mockCtrl)
	controller := New(mockRepo, nil)

	tests := []struct {
		name           string
		recordID       model.RecordID
		recordType     model.RecordType
		mockSetup      func()
		expectedResult float64
		expectedError  error
	}{
		{
			name:       "Successful retrieval of ratings",
			recordID:   "record1",
			recordType: "movie",
			mockSetup: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
						if recordID == "record1" && recordType == "movie" {
							return []model.Rating{{Value: 3}, {Value: 5}}, nil
						} else if recordID == "record2" && recordType == "book" {
							return nil, repository.ErrNotFound
						} else if recordID == "record3" && recordType == "music" {
							return nil, errors.New("database error")
						}
						return nil, errors.New("unexpected input")
					})

			},
			expectedResult: 4.0,
			expectedError:  nil,
		},
		{
			name:       "No ratings available",
			recordID:   "record4",
			recordType: "game",
			mockSetup: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, repository.ErrNotFound)
			},
			expectedResult: 0,
			expectedError:  ErrNotFound,
		},
		{
			name:       "Database error during retrieval",
			recordID:   "record5",
			recordType: "music",
			mockSetup: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
			},
			expectedResult: 0,
			expectedError:  errors.New("database error"),
		},
		{
			name:       "Invalid input",
			recordID:   "",
			recordType: "",
			mockSetup: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid input"))
			},
			expectedResult: 0,
			expectedError:  errors.New("invalid input"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := controller.GetAggregatedRating(context.Background(), tt.recordID, tt.recordType)
			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			}
		})
	}
}
func TestController_PutRating(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := gen.NewMockratingRepository(mockCtrl)
	controller := New(mockRepo, nil)

	tests := []struct {
		name          string
		recordID      model.RecordID
		recordType    model.RecordType
		rating        *model.Rating
		mockSetup     func()
		expectedError error
	}{
		{
			name:       "Successful rating put",
			recordID:   "record1",
			recordType: "movie",
			rating:     &model.Rating{UserID: "user1", Value: 5},
			mockSetup: func() {
				mockRepo.EXPECT().Put(gomock.Any(), gomock.Eq(model.RecordID("record1")), gomock.Eq(model.RecordType("movie")), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "Error during rating put",
			recordID:   "record2",
			recordType: "music",
			rating:     &model.Rating{UserID: "user2", Value: 3},
			mockSetup: func() {
				mockRepo.EXPECT().Put(gomock.Any(), gomock.Eq(model.RecordID("record2")), gomock.Eq(model.RecordType("music")), gomock.Any()).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := controller.PutRating(context.Background(), tt.recordID, tt.recordType, tt.rating)
			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestController_StartIngestion(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := gen.NewMockratingRepository(mockCtrl)
	mockIngester := gen.NewMockratingIngester(mockCtrl)
	controller := New(mockRepo, mockIngester)

	tests := []struct {
		name          string
		setupMocks    func()
		expectedError error
	}{
		{
			name: "Successful ingestion",
			setupMocks: func() {
				events := make(chan model.RatingEvent, 1)
				events <- model.RatingEvent{RecordID: "record1", RecordType: "movie", UserID: "user1", Value: 5}
				close(events)

				mockIngester.EXPECT().Ingest(gomock.Any()).Return(events, nil)
				mockRepo.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Ingestion error",
			setupMocks: func() {
				mockIngester.EXPECT().Ingest(gomock.Any()).Return(nil, errors.New("ingestion error"))
			},
			expectedError: errors.New("ingestion error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := controller.StartIngestion(context.Background())
			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			}
		})
	}
}
