package service

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"guidely-app/internal/utils"
	"time"
)

type AuthService struct {
	userRepo    *repository.UserRepo
	sessionRepo *repository.SessionRepo
}

func NewAuthService(userRepo *repository.UserRepo, sessionRepo *repository.SessionRepo) *AuthService {
	return &AuthService{userRepo: userRepo, sessionRepo: sessionRepo}
}

type RegisterInput struct {
	Nickname string
	Password string
}

type LoginInput struct {
	Nickname string
	Password string
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*models.User, string, error) {
	if !utils.IsValidNickname(input.Nickname) {
		return nil, "", errors.New("nickname must be at least 3 characters and max 50")
	}
	if len(input.Password) < 8 {
		return nil, "", errors.New("password must be at least 8 characters")
	}
	existing, _ := s.userRepo.GetByNickname(ctx, input.Nickname)
	if existing != nil {
		return nil, "", errors.New("nickname already exists")
	}
	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{
		Nickname:     input.Nickname,
		AvatarURL:    "",
		PasswordHash: hashed,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}
	token, err := utils.GenerateSessionToken()
	if err != nil {
		return nil, "", err
	}
	session := &models.Session{
		UserID:       user.ID,
		SessionToken: token,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*models.User, string, error) {
	user, err := s.userRepo.GetByNickname(ctx, input.Nickname)
	if err != nil || user == nil {
		return nil, "", errors.New("invalid credentials")
	}
	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		return nil, "", errors.New("invalid credentials")
	}
	token, err := utils.GenerateSessionToken()
	if err != nil {
		return nil, "", err
	}
	session := &models.Session{
		UserID:       user.ID,
		SessionToken: token,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteByToken(ctx, token)
}

func (s *AuthService) GetUserByID(ctx context.Context, id uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}
