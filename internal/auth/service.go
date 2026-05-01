package auth

import (
	"context"
	"errors"
	"time"

	"guidely-app/internal/auth/repository"
	"guidely-app/pkg/models"
	"guidely-app/pkg/utils"
)

type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (*models.User, string, error)
	Login(ctx context.Context, in LoginInput) (*models.User, string, error)
	Logout(ctx context.Context, token string) error
	GetUserByID(ctx context.Context, id uint64) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint64, nick, avatar, country, city, about *string) (*models.User, error)
	UpdateAvatar(ctx context.Context, id uint64, url string) (*models.User, error)
}

type RegisterInput struct{ Login, Password, Nickname string }
type LoginInput struct{ Login, Password string }

type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) AuthService {
	return &authService{userRepo: userRepo, sessionRepo: sessionRepo}
}

func (s *authService) Register(ctx context.Context, in RegisterInput) (*models.User, string, error) {
	if !utils.IsValidLogin(in.Login) {
		return nil, "", errors.New("invalid login format")
	}
	if !utils.IsValidNickname(in.Nickname) {
		return nil, "", errors.New("nickname must be 3..50 chars")
	}
	if len(in.Password) < 8 {
		return nil, "", errors.New("password too short")
	}
	if ex, _ := s.userRepo.GetByLogin(ctx, in.Login); ex != nil {
		return nil, "", errors.New("login already exists")
	}
	if exNick, _ := s.userRepo.GetByNickname(ctx, in.Nickname); exNick != nil {
		return nil, "", errors.New("nickname already exists")
	}
	hashed, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{Login: in.Login, Nickname: in.Nickname, PasswordHash: hashed}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}
	token, err := utils.GenerateSessionToken()
	if err != nil {
		return nil, "", err
	}
	session := &models.Session{
		UserID:    user.ID,
		TokenHash: utils.HashToken(token),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *authService) Login(ctx context.Context, in LoginInput) (*models.User, string, error) {
	user, err := s.userRepo.GetByLogin(ctx, in.Login)
	if err != nil || user == nil {
		return nil, "", errors.New("invalid credentials")
	}
	if !utils.CheckPasswordHash(in.Password, user.PasswordHash) {
		return nil, "", errors.New("invalid credentials")
	}
	token, err := utils.GenerateSessionToken()
	if err != nil {
		return nil, "", err
	}
	session := &models.Session{
		UserID:    user.ID,
		TokenHash: utils.HashToken(token),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteByToken(ctx, utils.HashToken(token))
}

func (s *authService) GetUserByID(ctx context.Context, id uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *authService) UpdateProfile(ctx context.Context, id uint64, nick, avatar, country, city, about *string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if nick != nil {
		user.Nickname = *nick
	}
	if avatar != nil {
		user.AvatarURL = *avatar
	}
	if country != nil {
		user.Country = country
	}
	if city != nil {
		user.City = city
	}
	if about != nil {
		user.About = about
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) UpdateAvatar(ctx context.Context, id uint64, url string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	user.AvatarURL = url
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
