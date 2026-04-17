package service

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"guidely-app/internal/utils"
	"time"
)

type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

type RegisterInput struct {
	Login    string
	Password string
	Nickname string
}

type LoginInput struct {
	Login    string
	Password string
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*models.User, string, error) {
	if !utils.IsValidLogin(input.Login) {
		return nil, "", errors.New("invalid login format")
	}
	if !utils.IsValidNickname(input.Nickname) {
		return nil, "", errors.New("nickname must be at least 3 characters and max 50")
	}
	if len(input.Password) < 8 {
		return nil, "", errors.New("password must be at least 8 characters")
	}
	existing, _ := s.userRepo.GetByLogin(ctx, input.Login)
	if existing != nil {
		return nil, "", errors.New("login already exists")
	}
	existingNick, _ := s.userRepo.GetByNickname(ctx, input.Nickname)
	if existingNick != nil {
		return nil, "", errors.New("nickname already exists")
	}
	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{
		Login:        input.Login,
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

func (s *authService) Login(ctx context.Context, input LoginInput) (*models.User, string, error) {
	user, err := s.userRepo.GetByLogin(ctx, input.Login)
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

func (s *authService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteByToken(ctx, token)
}

func (s *authService) GetUserByID(ctx context.Context, id uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}
