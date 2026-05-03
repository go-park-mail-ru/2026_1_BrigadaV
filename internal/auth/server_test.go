package auth

import (
	"context"
	"errors"
	"testing"
	pb "guidely-app/pkg/pb/auth"
	"guidely-app/pkg/models"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockAuthService struct {
	registerFn func(ctx context.Context, in RegisterInput) (*models.User, string, error)
}

func (m *mockAuthService) Register(ctx context.Context, in RegisterInput) (*models.User, string, error) {
	return m.registerFn(ctx, in)
}
func (m *mockAuthService) Login(ctx context.Context, in LoginInput) (*models.User, string, error) { return nil, "", nil }
func (m *mockAuthService) Logout(ctx context.Context, token string) error { return nil }
func (m *mockAuthService) GetUserByID(ctx context.Context, id uint64) (*models.User, error) { return nil, nil }
func (m *mockAuthService) UpdateProfile(ctx context.Context, id uint64, nick, avatar, country, city, about *string) (*models.User, error) { return nil, nil }
func (m *mockAuthService) UpdateAvatar(ctx context.Context, id uint64, url string) (*models.User, error) { return nil, nil }

func TestServer_Register_InvalidArgument(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(ctx context.Context, in RegisterInput) (*models.User, string, error) {
			return nil, "", errors.New("login already exists")
		},
	}
	srv := NewServer(svc)
	_, err := srv.Register(context.Background(), &pb.RegisterRequest{Login: "x", Password: "y", Nickname: "z"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}