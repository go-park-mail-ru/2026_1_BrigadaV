package service

import (
	"context"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"
)

type placeServiceImpl struct {
	placeRepo  repository.PlaceRepository
	reviewRepo repository.ReviewRepository
}

func NewPlaceService(placeRepo repository.PlaceRepository, reviewRepo repository.ReviewRepository) PlaceService {
	return &placeServiceImpl{placeRepo: placeRepo, reviewRepo: reviewRepo}
}

func (s *placeServiceImpl) GetAll(ctx context.Context) ([]models.Place, error) {
	return s.placeRepo.GetAll(ctx)
}

func (s *placeServiceImpl) GetByCategory(ctx context.Context, categoryID uint64) ([]models.Place, error) {
	return s.placeRepo.GetByCategory(ctx, categoryID)
}

func (s *placeServiceImpl) GetDetails(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error) {
	return s.placeRepo.GetWithRatingAndLike(ctx, placeID, userID)
}

func (s *placeServiceImpl) GetReviews(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
	return s.reviewRepo.GetByPlaceIDWithAuthor(ctx, placeID)
}

func (s *placeServiceImpl) IsPlaceInTrip(ctx context.Context, placeID, tripID uint64) (bool, error) {
	return s.placeRepo.IsPlaceInTrip(ctx, placeID, tripID)
}

func (s *placeServiceImpl) Search(ctx context.Context, query string) ([]models.Place, error) {
	return s.placeRepo.Search(ctx, query)
}
