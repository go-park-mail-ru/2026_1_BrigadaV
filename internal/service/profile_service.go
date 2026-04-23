package service

import (
	"context"
	"guidely-app/internal/logger"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"

	"github.com/sirupsen/logrus"
)

type profileService struct {
	userRepo repository.UserRepository
}

func NewProfileService(userRepo repository.UserRepository) ProfileService {
	return &profileService{userRepo: userRepo}
}

type UpdateProfileInput struct {
	Nickname  *string
	AvatarURL *string
	Country   *string
	City      *string
	About     *string
}

func (s *profileService) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *profileService) UpdateProfile(ctx context.Context, userID uint64, input UpdateProfileInput) (*models.User, error) {
	logger.Info(ctx, "UpdateProfile called", logrus.Fields{"user_id": userID})
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
	logger.Info(ctx, "UpdateProfile successful", logrus.Fields{"user_id": user.ID})
	return user, nil
}

func (s *profileService) UpdateAvatar(ctx context.Context, userID uint64, avatarURL string) (*models.User, error) {
	logger.Info(ctx, "UpdateAvatar called", logrus.Fields{"user_id": userID, "avatar_url": avatarURL})
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
