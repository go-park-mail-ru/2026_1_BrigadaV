package service

import (
	"context"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"
)

type countryServiceImpl struct {
	repo repository.CountryRepository
}

func NewCountryService(repo repository.CountryRepository) CountryService {
	return &countryServiceImpl{repo: repo}
}

func (s *countryServiceImpl) GetAll(ctx context.Context) ([]models.Country, error) {
	return s.repo.GetAll(ctx)
}

func (s *countryServiceImpl) GetWithLocalities(ctx context.Context, countryID uint64) (*models.Country, []models.Locality, error) {
	countries, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	var found *models.Country
	for _, c := range countries {
		if c.ID == countryID {
			c := c
			found = &c
			break
		}
	}
	if found == nil {
		return nil, nil, nil
	}
	localities, err := s.repo.GetLocalitiesByCountryID(ctx, countryID)
	if err != nil {
		return nil, nil, err
	}
	return found, localities, nil
}
