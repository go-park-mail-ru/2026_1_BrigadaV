package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

type PlaceService struct {
	placeRepo *repository.PlaceRepo
}

func NewPlaceService(placeRepo *repository.PlaceRepo) *PlaceService {
	return &PlaceService{placeRepo: placeRepo}
}

func (s *PlaceService) GetAll(ctx context.Context) ([]models.Place, error) {
	return s.placeRepo.GetAll(ctx)
}

func (s *PlaceService) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	return s.placeRepo.GetByID(ctx, id)
}
