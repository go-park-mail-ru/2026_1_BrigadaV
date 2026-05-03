package review

import (
	"context"
	"errors"
	"testing"

	"guidely-app/pkg/models"
	pb "guidely-app/pkg/pb/review"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockReviewService struct {
	createFn              func(ctx context.Context, input CreateReviewInput) (*models.Review, error)
	deleteFn              func(ctx context.Context, userID, reviewID uint64) error
	getByPlaceIDWithAuthorFn func(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error)
}

func (m *mockReviewService) Create(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return nil, nil
}

func (m *mockReviewService) Delete(ctx context.Context, userID, reviewID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, reviewID)
	}
	return nil
}

func (m *mockReviewService) GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
	if m.getByPlaceIDWithAuthorFn != nil {
		return m.getByPlaceIDWithAuthorFn(ctx, placeID)
	}
	return nil, nil
}

func TestServer_CreateReview_InvalidRating(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &mockReviewService{
		createFn: func(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
			return nil, errors.New("rating must be between 1 and 5")
		},
	}
	srv := NewServer(svc)

	_, err := srv.CreateReview(context.Background(), &pb.CreateReviewRequest{
		UserId:  1,
		PlaceId: 1,
		Rating:  6,
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}
