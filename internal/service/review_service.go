package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"guidely-app/internal/utils"
	"time"
)

type ReviewService struct {
	reviewRepo repository.ReviewRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo}
}

type CreateReviewInput struct {
	UserID    uint64
	PlaceID   uint64
	Rating    int16
	Comment   string
	VisitDate *time.Time
}

func (s *ReviewService) Create(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, utils.ErrBadRequest
	}
	review := &models.Review{
		UserID:    input.UserID,
		PlaceID:   input.PlaceID,
		Rating:    input.Rating,
		Comment:   input.Comment,
		VisitDate: input.VisitDate,
	}
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, utils.ErrInternal
	}
	return review, nil
}

func (s *ReviewService) GetByPlace(ctx context.Context, placeID uint64) ([]models.Review, error) {
	return s.reviewRepo.GetByPlaceID(ctx, placeID)
}

func (s *ReviewService) Delete(ctx context.Context, userID, reviewID uint64) error {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil || review == nil {
		return utils.ErrNotFound
	}
	if review.UserID != userID {
		return utils.ErrUnauthorized
	}
	return s.reviewRepo.Delete(ctx, reviewID)
}
