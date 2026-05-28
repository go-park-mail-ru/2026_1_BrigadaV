// package review

// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"

// 	"guidely-app/pkg/models"
// 	pb "guidely-app/pkg/pb/review"

// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// type mockReviewService struct {
// 	createFn                 func(ctx context.Context, input CreateReviewInput) (*models.Review, error)
// 	deleteFn                 func(ctx context.Context, userID, reviewID uint64) error
// 	getByPlaceIDWithAuthorFn func(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error)
// }

// func (m *mockReviewService) Create(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
// 	if m.createFn != nil {
// 		return m.createFn(ctx, input)
// 	}
// 	return nil, nil
// }

// func (m *mockReviewService) Delete(ctx context.Context, userID, reviewID uint64) error {
// 	if m.deleteFn != nil {
// 		return m.deleteFn(ctx, userID, reviewID)
// 	}
// 	return nil
// }

// func (m *mockReviewService) GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
// 	if m.getByPlaceIDWithAuthorFn != nil {
// 		return m.getByPlaceIDWithAuthorFn(ctx, placeID)
// 	}
// 	return nil, nil
// }

// func TestServer_CreateReview_InvalidRating(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	svc := &mockReviewService{
// 		createFn: func(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
// 			return nil, errors.New("rating must be between 1 and 5")
// 		},
// 	}
// 	srv := NewServer(svc)

// 	_, err := srv.CreateReview(context.Background(), &pb.CreateReviewRequest{
// 		UserId:  1,
// 		PlaceId: 1,
// 		Rating:  6,
// 	})
// 	st, _ := status.FromError(err)
// 	assert.Equal(t, codes.Internal, st.Code())
// }

// func TestServer_GetReviewsByPlace(t *testing.T) {
// 	svc := &mockReviewService{
// 		getByPlaceIDWithAuthorFn: func(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
// 			return []models.ReviewWithAuthor{{ID: 1}}, nil
// 		},
// 	}
// 	srv := NewServer(svc)
// 	resp, err := srv.GetReviewsByPlace(context.Background(), &pb.GetReviewsByPlaceRequest{PlaceId: 1})
// 	assert.NoError(t, err)
// 	assert.Len(t, resp.Reviews, 1)
// }

// func TestServer_CreateReview_DBError(t *testing.T) {
// 	svc := &mockReviewService{
// 		createFn: func(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
// 			return nil, errors.New("db error")
// 		},
// 	}
// 	srv := NewServer(svc)
// 	_, err := srv.CreateReview(context.Background(), &pb.CreateReviewRequest{UserId: 1, PlaceId: 1, Rating: 5})
// 	st, _ := status.FromError(err)
// 	assert.Equal(t, codes.Internal, st.Code())
// }

// func TestServer_DeleteReview_NotFound(t *testing.T) {
// 	svc := &mockReviewService{
// 		deleteFn: func(ctx context.Context, userID, reviewID uint64) error { return errors.New("review not found") },
// 	}
// 	srv := NewServer(svc)
// 	_, err := srv.DeleteReview(context.Background(), &pb.DeleteReviewRequest{UserId: 1, ReviewId: 1})
// 	st, _ := status.FromError(err)
// 	assert.Equal(t, codes.Internal, st.Code())
// }

// // TestServer_CreateReview_AlreadyExists проверяет что при дублирующемся отзыве
// // возвращается codes.AlreadyExists, а не codes.Internal.
// func TestServer_CreateReview_AlreadyExists(t *testing.T) {
// 	svc := &mockReviewService{
// 		createFn: func(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
// 			// Симулируем ошибку unique constraint от PostgreSQL
// 			return nil, errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
// 		},
// 	}
// 	srv := NewServer(svc)
// 	_, err := srv.CreateReview(context.Background(), &pb.CreateReviewRequest{UserId: 1, PlaceId: 1, Rating: 4})
// 	st, _ := status.FromError(err)
// 	assert.Equal(t, codes.AlreadyExists, st.Code())
// 	assert.Contains(t, st.Message(), "already reviewed")
// }

// // TestServer_CreateReview_Success проверяет успешное создание отзыва.
// func TestServer_CreateReview_Success(t *testing.T) {
// 	svc := &mockReviewService{
// 		createFn: func(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
// 			return &models.Review{ID: 42, UserID: 1, PlaceID: 1, Rating: 5}, nil
// 		},
// 	}
// 	srv := NewServer(svc)
// 	resp, err := srv.CreateReview(context.Background(), &pb.CreateReviewRequest{UserId: 1, PlaceId: 1, Rating: 5})
// 	assert.NoError(t, err)
// 	assert.Equal(t, uint64(42), resp.Id)
// }

// // TestServer_DeleteReview_Success проверяет успешное удаление отзыва.
// func TestServer_DeleteReview_Success(t *testing.T) {
// 	svc := &mockReviewService{
// 		deleteFn: func(ctx context.Context, userID, reviewID uint64) error { return nil },
// 	}
// 	srv := NewServer(svc)
// 	_, err := srv.DeleteReview(context.Background(), &pb.DeleteReviewRequest{UserId: 1, ReviewId: 1})
// 	assert.NoError(t, err)
// }

// // TestServer_GetReviewsByPlace_Error проверяет обработку ошибки при получении отзывов.
// func TestServer_GetReviewsByPlace_Error(t *testing.T) {
// 	svc := &mockReviewService{
// 		getByPlaceIDWithAuthorFn: func(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
// 			return nil, errors.New("db error")
// 		},
// 	}
// 	srv := NewServer(svc)
// 	_, err := srv.GetReviewsByPlace(context.Background(), &pb.GetReviewsByPlaceRequest{PlaceId: 1})
// 	st, _ := status.FromError(err)
// 	assert.Equal(t, codes.Internal, st.Code())
// }

// // TestToReviewResponse проверяет функцию toReviewResponse
// func TestToReviewResponse(t *testing.T) {
// 	title := "Great"
// 	visitDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
// 	r := &models.Review{
// 		ID:        1,
// 		UserID:    2,
// 		PlaceID:   3,
// 		Title:     &title,
// 		Rating:    5,
// 		Comment:   "Excellent",
// 		VisitDate: &visitDate,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}
// 	resp := toReviewResponse(r)
// 	assert.Equal(t, uint64(1), resp.Id)
// 	assert.Equal(t, uint64(2), resp.UserId)
// 	assert.Equal(t, uint64(3), resp.PlaceId)
// 	assert.Equal(t, "Great", *resp.Title)
// 	assert.Equal(t, int32(5), resp.Rating)
// 	assert.Equal(t, "Excellent", resp.Comment)
// 	assert.NotNil(t, resp.VisitDate)
// 	assert.Equal(t, "2024-01-15", *resp.VisitDate)
// }

// // TestToReviewWithAuthorMessages проверяет функцию toReviewWithAuthorMessages

// // TestIsDuplicateError проверяет функцию isDuplicateError
// func TestIsDuplicateError(t *testing.T) {
// 	assert.True(t, isDuplicateError(errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")))
// 	assert.True(t, isDuplicateError(errors.New("unique constraint violation")))
// 	assert.False(t, isDuplicateError(errors.New("some other error")))
// }

// // TestServer_CreateReview_Duplicate проверяет что при дубликате возвращается AlreadyExists
// func TestServer_CreateReview_Duplicate(t *testing.T) {
// 	svc := &mockReviewService{
// 		createFn: func(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
// 			return nil, errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
// 		},
// 	}
// 	srv := NewServer(svc)
// 	_, err := srv.CreateReview(context.Background(), &pb.CreateReviewRequest{UserId: 1, PlaceId: 1, Rating: 4})
// 	st, _ := status.FromError(err)
// 	assert.Equal(t, codes.AlreadyExists, st.Code())
// }

// // TestServer_CreateReview_VisitDateParseError проверяет случай неверной даты
// func TestServer_CreateReview_VisitDateParseError(t *testing.T) {
// 	svc := &mockReviewService{
// 		createFn: func(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
// 			// Если дата не распарсилась, она должна быть nil
// 			assert.Nil(t, input.VisitDate)
// 			return &models.Review{ID: 1}, nil
// 		},
// 	}
// 	srv := NewServer(svc)
// 	invalidDate := "invalid-date"
// 	resp, err := srv.CreateReview(context.Background(), &pb.CreateReviewRequest{UserId: 1, PlaceId: 1, Rating: 5, VisitDate: &invalidDate})
// 	assert.NoError(t, err)
// 	assert.Equal(t, uint64(1), resp.Id)
// }
