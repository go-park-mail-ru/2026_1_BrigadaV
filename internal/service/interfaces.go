package service

import (
	"context"
	"guidely-app/pkg/models"
)

type PlaceService interface {
	GetAll(ctx context.Context) ([]models.Place, error)
	GetDetails(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error)
	GetReviews(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error)
	IsPlaceInTrip(ctx context.Context, placeID, tripID uint64) (bool, error)
}

type ProfileService interface {
	GetProfile(ctx context.Context, userID uint64) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uint64, input UpdateProfileInput) (*models.User, error)
	UpdateAvatar(ctx context.Context, userID uint64, avatarURL string) (*models.User, error)
}

type TripService interface {
	Create(ctx context.Context, input CreateTripInput) (*models.Trip, error)
	GetUserTrips(ctx context.Context, userID uint64) ([]models.Trip, error)
	GetTripDetails(ctx context.Context, tripID uint64) (*models.Trip, []models.PlaceInTrip, error)
	Update(ctx context.Context, id, userID uint64, input UpdateTripInput) (*models.Trip, error)
	Delete(ctx context.Context, id, userID uint64) error
	GetTripPlaceIDs(ctx context.Context, tripID uint64) ([]uint64, error)
	AddPlaceToTrip(ctx context.Context, tripID, placeID, userID uint64, orderIndex int16) error
	RemovePlaceFromTrip(ctx context.Context, tripID, placeID, userID uint64) error
}

type ReviewService interface {
	Create(ctx context.Context, input CreateReviewInput) (*models.Review, error)
	Delete(ctx context.Context, userID, reviewID uint64) error
}

type CategoryService interface {
	GetAll(ctx context.Context) ([]models.Category, error)
	GetByID(ctx context.Context, id uint64) (*models.Category, error)
	Create(ctx context.Context, c *models.Category) error
	Update(ctx context.Context, c *models.Category) error
	Delete(ctx context.Context, id uint64) error
}
