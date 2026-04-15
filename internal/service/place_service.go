package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

type PlaceService struct {
	placeRepo  *repository.PlaceRepo
	reviewRepo *repository.ReviewRepo
}

func NewPlaceService(placeRepo *repository.PlaceRepo, reviewRepo *repository.ReviewRepo) *PlaceService {
	return &PlaceService{placeRepo: placeRepo, reviewRepo: reviewRepo}
}

func (s *PlaceService) GetAll(ctx context.Context) ([]models.Place, error) {
	return s.placeRepo.GetAll(ctx)
}

func (s *PlaceService) GetDetails(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error) {
	return s.placeRepo.GetWithRatingAndLike(ctx, placeID, userID)
}

func (s *PlaceService) GetReviews(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
	return s.reviewRepo.GetByPlaceIDWithAuthor(ctx, placeID)
}
