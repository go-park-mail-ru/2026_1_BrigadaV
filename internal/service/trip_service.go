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
	Title      string
	Location   *string
	StartDate  *time.Time
	EndDate    *time.Time
	PreviewURL *string
	CreatedBy  uint64
	IsPublic   bool
}

type UpdateTripInput struct {
	Title       *string
	Description *string
	Location    *string
	StartDate   *time.Time
	EndDate     *time.Time
	PreviewURL  *string
	IsPublic    *bool
}

func (s *TripService) Create(ctx context.Context, input CreateTripInput) (*models.Trip, error) {
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	trip := &models.Trip{
		Title:      input.Title,
		Location:   input.Location,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		PreviewURL: input.PreviewURL,
		CreatedBy:  input.CreatedBy,
		IsPublic:   input.IsPublic,
	}
	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *TripService) GetUserTrips(ctx context.Context, userID uint64) ([]models.Trip, error) {
	return s.tripRepo.GetByUser(ctx, userID)
}

func (s *TripService) GetTripDetails(ctx context.Context, tripID uint64) (*models.Trip, []models.PlaceInTrip, error) {
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil || trip == nil {
		return nil, nil, errors.New("trip not found")
	}
	places, err := s.tripRepo.GetAttractions(ctx, tripID)
	if err != nil {
		return nil, nil, err
	}
	return trip, places, nil
}

func (s *TripService) Update(ctx context.Context, id, userID uint64, input UpdateTripInput) (*models.Trip, error) {
	trip, err := s.tripRepo.GetByID(ctx, id)
	if err != nil || trip == nil {
		return nil, errors.New("trip not found")
	}
	if trip.CreatedBy != userID {
		return nil, errors.New("not authorized")
	}
	if input.Title != nil {
		trip.Title = *input.Title
	}
	if input.Description != nil {
		trip.Description = *input.Description
	}
	if input.Location != nil {
		trip.Location = input.Location
	}
	if input.StartDate != nil {
		trip.StartDate = input.StartDate
	}
	if input.EndDate != nil {
		trip.EndDate = input.EndDate
	}
	if input.PreviewURL != nil {
		trip.PreviewURL = input.PreviewURL
	}
	if input.IsPublic != nil {
		trip.IsPublic = *input.IsPublic
	}
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
		return errors.New("not authorized")
	}
	return s.tripRepo.Delete(ctx, id)
}
