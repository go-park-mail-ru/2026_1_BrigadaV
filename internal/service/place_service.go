package service

import (
	"context"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"
)

type placeServiceImpl struct {
	placeRepo repository.PlaceRepository
}

func NewPlaceService(placeRepo repository.PlaceRepository) PlaceService {
	return &placeServiceImpl{placeRepo: placeRepo}
}

func (s *placeServiceImpl) GetAll(ctx context.Context) ([]models.Place, error) {
	return s.placeRepo.GetAll(ctx)
}

func (s *placeServiceImpl) GetDetails(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error) {
	return s.placeRepo.GetWithRatingAndLike(ctx, placeID, userID)
}
