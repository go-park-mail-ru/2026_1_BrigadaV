package auth

import (
	"context"
	"errors"
	"testing"

	"guidely-app/pkg/models"
	pb "guidely-app/pkg/pb/auth"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockAuthService struct {
	registerFn      func(ctx context.Context, in RegisterInput) (*models.User, string, error)
	loginFn         func(ctx context.Context, in LoginInput) (*models.User, string, error)
	logoutFn        func(ctx context.Context, token string) error
	getUserByIDFn   func(ctx context.Context, id uint64) (*models.User, error)
	updateProfileFn func(ctx context.Context, id uint64, nick, avatar, country, city, about *string) (*models.User, error)
	updateAvatarFn  func(ctx context.Context, id uint64, url string) (*models.User, error)
}

func (m *mockAuthService) Register(ctx context.Context, in RegisterInput) (*models.User, string, error) {
	return m.registerFn(ctx, in)
}
func (m *mockAuthService) Login(ctx context.Context, in LoginInput) (*models.User, string, error) {
	return m.loginFn(ctx, in)
}
func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	return m.logoutFn(ctx, token)
}
func (m *mockAuthService) GetUserByID(ctx context.Context, id uint64) (*models.User, error) {
	return m.getUserByIDFn(ctx, id)
}
func (m *mockAuthService) UpdateProfile(ctx context.Context, id uint64, nick, avatar, country, city, about *string) (*models.User, error) {
	return m.updateProfileFn(ctx, id, nick, avatar, country, city, about)
}
func (m *mockAuthService) UpdateAvatar(ctx context.Context, id uint64, url string) (*models.User, error) {
	return m.updateAvatarFn(ctx, id, url)
}

func TestServer_Register_InvalidArgument(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(ctx context.Context, in RegisterInput) (*models.User, string, error) {
			return nil, "", errors.New("login already exists")
		},
	}
	srv := NewServer(svc)
	_, err := srv.Register(context.Background(), &pb.RegisterRequest{
		Login: "test@example.com", Password: "12345678", Nickname: "tester",
	})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestServer_Login_InvalidCredentials(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(ctx context.Context, in LoginInput) (*models.User, string, error) {
			return nil, "", errors.New("invalid credentials")
		},
	}
	srv := NewServer(svc)
	_, err := srv.Login(context.Background(), &pb.LoginRequest{Login: "test", Password: "wrong"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestServer_Logout_Error(t *testing.T) {
	svc := &mockAuthService{
		logoutFn: func(ctx context.Context, token string) error {
			return errors.New("db error")
		},
	}
	srv := NewServer(svc)
	_, err := srv.Logout(context.Background(), &pb.LogoutRequest{Token: "token123"})
	assert.Error(t, err)
}

func TestServer_GetUser_NotFound(t *testing.T) {
	svc := &mockAuthService{
		getUserByIDFn: func(ctx context.Context, id uint64) (*models.User, error) {
			return nil, errors.New("not found")
		},
	}
	srv := NewServer(svc)
	_, err := srv.GetUser(context.Background(), &pb.GetUserRequest{UserId: 1})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestServer_UpdateProfile_NotFound(t *testing.T) {
	svc := &mockAuthService{
		updateProfileFn: func(ctx context.Context, id uint64, nick, avatar, country, city, about *string) (*models.User, error) {
			return nil, errors.New("not found")
		},
	}
	srv := NewServer(svc)
	_, err := srv.UpdateProfile(context.Background(), &pb.UpdateProfileRequest{UserId: 1, Nickname: ptrString("new")})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func ptrString(s string) *string { return &s }
