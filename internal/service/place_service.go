package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

type PlaceService struct {
	placeRepo repository.PlaceRepository
}

func NewPlaceService(placeRepo repository.PlaceRepository) *PlaceService {
	return &PlaceService{placeRepo: placeRepo}
}

func (s *PlaceService) GetAll(ctx context.Context) ([]models.Place, error) {
	return s.placeRepo.GetAll(ctx)
}

func (s *PlaceService) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	return s.placeRepo.GetByID(ctx, id)
}
