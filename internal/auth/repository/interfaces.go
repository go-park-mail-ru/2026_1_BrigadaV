// internal/auth/repository/interfaces.go
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
	GetByYandexID(ctx context.Context, yandexID string) (*models.User, error)
	CreateOAuth(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	DeleteByToken(ctx context.Context, token string) error
}
