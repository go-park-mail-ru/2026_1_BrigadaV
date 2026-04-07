package service

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"time"
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepo
}

func NewReviewService(reviewRepo *repository.ReviewRepo) *ReviewService {
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
		return nil, errors.New("rating must be between 1 and 5")
	}
	review := &models.Review{
		UserID:    input.UserID,
		PlaceID:   input.PlaceID,
		Rating:    input.Rating,
		Comment:   input.Comment,
		VisitDate: input.VisitDate,
	}
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *ReviewService) GetByPlace(ctx context.Context, placeID uint64) ([]models.Review, error) {
	return s.reviewRepo.GetByPlaceID(ctx, placeID)
}

func (s *ReviewService) Delete(ctx context.Context, userID, reviewID uint64) error {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil || review == nil {
		return errors.New("review not found")
	}
	if review.UserID != userID {
		return errors.New("not authorized to delete this review")
	}
	return s.reviewRepo.Delete(ctx, reviewID)
}
