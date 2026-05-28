package service

import (
	"context"
	"errors"

	"guidely-app/internal/repository"
	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
)

type UpdateProfileInput struct {
	Nickname  *string
	AvatarURL *string
	Country   *string
	City      *string
	About     *string
}

type profileServiceImpl struct {
	userRepo repository.UserRepository
}

func NewProfileService(userRepo repository.UserRepository) ProfileService {
	return &profileServiceImpl{userRepo: userRepo}
}

func (s *profileServiceImpl) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *profileServiceImpl) UpdateProfile(ctx context.Context, userID uint64, input UpdateProfileInput) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Проверка уникальности nickname
	if input.Nickname != nil && *input.Nickname != user.Nickname {
		existing, _ := s.userRepo.GetByNickname(ctx, *input.Nickname)
		if existing != nil && existing.ID != userID {
			return nil, errors.New("nickname already taken")
		}
		user.Nickname = *input.Nickname
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}
	if input.Country != nil {
		user.Country = input.Country
	}
	if input.City != nil {
		user.City = input.City
	}
	if input.About != nil {
		user.About = input.About
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateAvatar – обновляет аватар пользователя с использованием транзакции и блокировки строки
func (s *profileServiceImpl) UpdateAvatar(ctx context.Context, userID uint64, avatarURL string) (*models.User, error) {
	// Получаем адаптер, который реализует интерфейс DB с Begin
	dbAdapter, ok := s.userRepo.(interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	})
	if !ok {
		// fallback на старую логику, если Begin не поддерживается (для тестов)
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, errors.New("user not found")
		}
		user.AvatarURL = avatarURL
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	tx, err := dbAdapter.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Получаем пользователя с блокировкой строки
	user, err := s.userRepo.(interface {
		GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uint64) (*models.User, error)
	}).GetByIDForUpdate(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Обновляем аватар
	user.AvatarURL = avatarURL

	// Обновляем в БД в рамках транзакции
	if err := s.userRepo.(interface {
		UpdateWithTx(ctx context.Context, tx pgx.Tx, user *models.User) error
	}).UpdateWithTx(ctx, tx, user); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return user, nil
}
