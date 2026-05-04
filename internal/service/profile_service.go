package service

import (
	"context"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"
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
	if input.Nickname != nil {
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

func (s *profileServiceImpl) UpdateAvatar(ctx context.Context, userID uint64, avatarURL string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.AvatarURL = avatarURL
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
