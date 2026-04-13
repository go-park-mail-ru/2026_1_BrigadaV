package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

type ProfileService struct {
	userRepo repository.UserRepository
}

func NewProfileService(userRepo repository.UserRepository) *ProfileService {
	return &ProfileService{userRepo: userRepo}
}

type UpdateProfileInput struct {
	Nickname  *string
	AvatarURL *string
	Country   *string
	City      *string
	About     *string
}

func (s *ProfileService) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID uint64, input UpdateProfileInput) (*models.User, error) {
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
