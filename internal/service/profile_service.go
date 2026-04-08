package service

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"guidely-app/internal/utils"
)

type ProfileService struct {
	userRepo repository.UserRepository
}

func NewProfileService(userRepo repository.UserRepository) *ProfileService {
	return &ProfileService{userRepo: userRepo}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID uint64, nickname, avatarURL string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if nickname != "" {
		user.Nickname = nickname
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.New("failed to update profile")
	}
	return user, nil
}

func (s *ProfileService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}
	if !utils.CheckPasswordHash(oldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}
	user.PasswordHash = hashedPassword
	return s.userRepo.Update(ctx, user)
}
