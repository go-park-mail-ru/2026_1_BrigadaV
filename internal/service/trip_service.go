package service

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"time"
)

type tripService struct {
	tripRepo repository.TripRepository
}

func NewTripService(tripRepo repository.TripRepository) TripService {
	return &tripService{tripRepo: tripRepo}
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

func (s *tripService) Create(ctx context.Context, input CreateTripInput) (*models.Trip, error) {
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

func (s *tripService) GetUserTrips(ctx context.Context, userID uint64) ([]models.Trip, error) {
	return s.tripRepo.GetByUser(ctx, userID)
}

func (s *tripService) GetTripDetails(ctx context.Context, tripID uint64) (*models.Trip, []models.PlaceInTrip, error) {
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, nil, err
	}
	if trip == nil {
		return nil, nil, errors.New("trip not found")
	}
	places, err := s.tripRepo.GetAttractions(ctx, tripID)
	if err != nil {
		return nil, nil, err
	}
	return trip, places, nil
}

func (s *tripService) Update(ctx context.Context, id, userID uint64, input UpdateTripInput) (*models.Trip, error) {
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

func (s *tripService) Delete(ctx context.Context, id, userID uint64) error {
	trip, err := s.tripRepo.GetByID(ctx, id)
	if err != nil || trip == nil {
		return errors.New("trip not found")
	}
	if trip.CreatedBy != userID {
		return errors.New("not authorized")
	}
	return s.tripRepo.Delete(ctx, id)
}

func (s *tripService) GetTripPlaceIDs(ctx context.Context, tripID uint64) ([]uint64, error) {
	return s.tripRepo.GetPlaceIDs(ctx, tripID)
}

func (s *tripService) AddPlaceToTrip(ctx context.Context, tripID, placeID, userID uint64, orderIndex int16) error {
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil || trip == nil {
		return errors.New("trip not found")
	}
	if trip.CreatedBy != userID {
		return errors.New("not your trip")
	}

	exists, err := s.tripRepo.CheckPlaceInTrip(ctx, tripID, placeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("place already in trip")
	}

	return s.tripRepo.AddAttraction(ctx, tripID, placeID, orderIndex)
}

func (s *tripService) RemovePlaceFromTrip(ctx context.Context, tripID, placeID, userID uint64) error {
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil || trip == nil {
		return errors.New("trip not found")
	}
	if trip.CreatedBy != userID {
		return errors.New("not your trip")
	}
	return s.tripRepo.RemoveAttraction(ctx, tripID, placeID)
}
