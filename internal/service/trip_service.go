package service

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"time"
)

type TripService struct {
	tripRepo *repository.TripRepo
}

func NewTripService(tripRepo *repository.TripRepo) *TripService {
	return &TripService{tripRepo: tripRepo}
}

type CreateTripInput struct {
	Title       string
	Description string
	StartDate   *time.Time
	EndDate     *time.Time
	CreatedBy   uint64
	IsPublic    bool
}

type UpdateTripInput struct {
	Title       string
	Description string
	StartDate   *time.Time
	EndDate     *time.Time
	IsPublic    bool
}

func (s *TripService) Create(ctx context.Context, input CreateTripInput) (*models.Trip, error) {
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	trip := &models.Trip{
		Title:       input.Title,
		Description: input.Description,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		CreatedBy:   input.CreatedBy,
		IsPublic:    input.IsPublic,
	}
	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *TripService) GetByID(ctx context.Context, id uint64) (*models.Trip, error) {
	return s.tripRepo.GetByID(ctx, id)
}

func (s *TripService) GetUserTrips(ctx context.Context, userID uint64) ([]models.Trip, error) {
	return s.tripRepo.GetByUser(ctx, userID)
}

func (s *TripService) Update(ctx context.Context, id, userID uint64, input UpdateTripInput) (*models.Trip, error) {
	trip, err := s.tripRepo.GetByID(ctx, id)
	if err != nil || trip == nil {
		return nil, errors.New("trip not found")
	}
	if trip.CreatedBy != userID {
		return nil, errors.New("not authorized to update this trip")
	}
	if input.Title != "" {
		trip.Title = input.Title
	}
	if input.Description != "" {
		trip.Description = input.Description
	}
	trip.StartDate = input.StartDate
	trip.EndDate = input.EndDate
	trip.IsPublic = input.IsPublic
	if err := s.tripRepo.Update(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *TripService) Delete(ctx context.Context, id, userID uint64) error {
	trip, err := s.tripRepo.GetByID(ctx, id)
	if err != nil || trip == nil {
		return errors.New("trip not found")
	}
	if trip.CreatedBy != userID {
		return errors.New("not authorized to delete this trip")
	}
	return s.tripRepo.Delete(ctx, id)
}
