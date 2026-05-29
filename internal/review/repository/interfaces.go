package repository

import (
	"context"
	"guidely-app/pkg/models"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *models.Review) error
	GetByID(ctx context.Context, id uint64) (*models.Review, error)
	GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error)
	Delete(ctx context.Context, id uint64) error
	ExistsByUserAndPlace(ctx context.Context, userID, placeID uint64) (bool, error)
}
