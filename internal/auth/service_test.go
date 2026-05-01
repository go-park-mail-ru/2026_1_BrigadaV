package auth

import (
	"context"
	"testing"

	"guidely-app/internal/auth/repository/mocks"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	userRepo.EXPECT().GetByLogin(gomock.Any(), "test@example.com").Return(nil, nil)
	userRepo.EXPECT().GetByNickname(gomock.Any(), "tester").Return(nil, nil)
	userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *models.User) error {
		u.ID = 1
		return nil
	})
	sessRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	user, token, err := svc.Register(context.Background(), RegisterInput{
		Login: "test@example.com", Password: "12345678", Nickname: "tester",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, uint64(1), user.ID)
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	userRepo.EXPECT().GetByLogin(gomock.Any(), "test@example.com").Return(nil, nil)

	user, token, err := svc.Login(context.Background(), LoginInput{
		Login: "test@example.com", Password: "wrong",
	})
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
}
