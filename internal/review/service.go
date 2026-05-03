package review

import (
	"context"
	"errors"
	"time"

	"guidely-app/internal/review/repository"
	"guidely-app/pkg/models"
)

type ReviewService interface {
	Create(ctx context.Context, input CreateReviewInput) (*models.Review, error)
	Delete(ctx context.Context, userID, reviewID uint64) error
	GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error)
}

type CreateReviewInput struct {
	UserID    uint64
	PlaceID   uint64
	Title     *string
	Rating    int16
	Comment   string
	VisitDate *time.Time
}

type reviewServiceImpl struct {
	repo repository.ReviewRepository
}

func NewService(repo repository.ReviewRepository) ReviewService {
	return &reviewServiceImpl{repo: repo}
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
	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *reviewServiceImpl) Delete(ctx context.Context, userID, reviewID uint64) error {
	review, err := s.repo.GetByID(ctx, reviewID)
	if err != nil || review == nil {
		return errors.New("review not found")
	}
	if review.UserID != userID {
		return errors.New("not authorized")
	}
	return s.repo.Delete(ctx, reviewID)
}

func (s *reviewServiceImpl) GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
	return s.repo.GetByPlaceIDWithAuthor(ctx, placeID)
}
