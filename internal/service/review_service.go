package service

import (
	"context"
	"errors"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"
	"time"
)

type reviewServiceImpl struct {
	reviewRepo repository.ReviewRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository) ReviewService {
	return &reviewServiceImpl{reviewRepo: reviewRepo}
}

type CreateReviewInput struct {
	UserID    uint64
	PlaceID   uint64
	Title     *string
	Rating    int16
	Comment   string
	VisitDate *time.Time
}

func (s *reviewServiceImpl) Create(ctx context.Context, input CreateReviewInput) (*models.Review, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	review := &models.Review{
		UserID:    input.UserID,
		PlaceID:   input.PlaceID,
		Title:     input.Title,
		Rating:    input.Rating,
		Comment:   input.Comment,
		VisitDate: input.VisitDate,
	}
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *reviewServiceImpl) Delete(ctx context.Context, userID, reviewID uint64) error {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil || review == nil {
		return errors.New("review not found")
	}
	if review.UserID != userID {
		return errors.New("not authorized")
	}
	return s.reviewRepo.Delete(ctx, reviewID)
}
