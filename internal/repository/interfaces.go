package repository

import (
	"context"
	"guidely-app/pkg/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)
	GetByID(ctx context.Context, id uint64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	DeleteByToken(ctx context.Context, token string) error
}

type PlaceRepository interface {
	GetAll(ctx context.Context) ([]models.Place, error)
	GetByID(ctx context.Context, id uint64) (*models.Place, error)
	GetByIDs(ctx context.Context, ids []uint64) ([]models.Place, error)
	GetWithRatingAndLike(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error)
	IsPlaceInTrip(ctx context.Context, placeID, tripID uint64) (bool, error)
	Search(ctx context.Context, query string) ([]models.Place, error)
	GetByCategory(ctx context.Context, categoryID uint64) ([]models.Place, error)
}

// PlaceSearchRepository — интерфейс для полнотекстового поиска (ElasticSearch).
// Позволяет подменить реализацию в тестах или откатиться на SQL при недоступности ES.
type PlaceSearchRepository interface {
	Search(ctx context.Context, query string) ([]models.Place, error)
}

type TripRepository interface {
	Create(ctx context.Context, trip *models.Trip) error
	GetByID(ctx context.Context, id uint64) (*models.Trip, error)
	GetByUser(ctx context.Context, userID uint64) ([]models.Trip, error)
	Update(ctx context.Context, trip *models.Trip) error
	Delete(ctx context.Context, id uint64) error
	AddAttraction(ctx context.Context, tripID, placeID uint64, order int16) error
	GetAttractions(ctx context.Context, tripID uint64) ([]models.PlaceInTrip, error)
	GetPlaceIDs(ctx context.Context, tripID uint64) ([]uint64, error)
	RemoveAttraction(ctx context.Context, tripID, placeID uint64) error
	CheckPlaceInTrip(ctx context.Context, tripID, placeID uint64) (bool, error)
}

type CategoryRepository interface {
	GetAll(ctx context.Context) ([]models.Category, error)
	GetByID(ctx context.Context, id uint64) (*models.Category, error)
	Create(ctx context.Context, c *models.Category) error
	Update(ctx context.Context, c *models.Category) error
	Delete(ctx context.Context, id uint64) error
}

type ReviewRepository interface {
	Create(ctx context.Context, review *models.Review) error
	GetByID(ctx context.Context, id uint64) (*models.Review, error)
	GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error)
	Delete(ctx context.Context, id uint64) error
}
