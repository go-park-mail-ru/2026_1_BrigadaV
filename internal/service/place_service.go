package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

type placeService struct {
	placeRepo  repository.PlaceRepository
	reviewRepo repository.ReviewRepository
}

func NewPlaceService(placeRepo repository.PlaceRepository, reviewRepo repository.ReviewRepository) PlaceService {
	return &placeService{
		placeRepo:  placeRepo,
		reviewRepo: reviewRepo,
	}
}

func (s *placeService) GetAll(ctx context.Context) ([]models.Place, error) {
	return s.placeRepo.GetAll(ctx)
}

func (s *placeService) GetDetails(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error) {
	return s.placeRepo.GetWithRatingAndLike(ctx, placeID, userID)
}

func (s *placeService) GetReviews(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
	return s.reviewRepo.GetByPlaceIDWithAuthor(ctx, placeID)
}

func (s *placeService) IsPlaceInTrip(ctx context.Context, placeID, tripID uint64) (bool, error) {
	return s.placeRepo.IsPlaceInTrip(ctx, placeID, tripID)
}

func (s *placeService) Search(ctx context.Context, query string) ([]models.Place, error) {
	return s.placeRepo.Search(ctx, query)
}
