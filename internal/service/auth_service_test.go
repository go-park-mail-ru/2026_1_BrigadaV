package service

import (
	"context"
	"testing"

	"guidely-app/internal/models"
	"guidely-app/internal/repository/mocks"
	"guidely-app/internal/utils"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	service := NewAuthService(mockUserRepo, mockSessionRepo)

	mockUserRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, nil)
	mockUserRepo.EXPECT().GetByNickname(gomock.Any(), "tester").Return(nil, nil)
	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockSessionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	user, token, err := service.Register(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "12345678",
		Nickname: "tester",
	})

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	service := NewAuthService(mockUserRepo, mockSessionRepo)

	existingUser := &models.User{ID: 1, Email: "test@example.com"}
	mockUserRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(existingUser, nil)

	user, token, err := service.Register(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "12345678",
		Nickname: "tester",
	})

	assert.Error(t, err)
	assert.Equal(t, "email already exists", err.Error())
	assert.Nil(t, user)
	assert.Empty(t, token)
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	service := NewAuthService(mockUserRepo, mockSessionRepo)

	user, token, err := service.Register(context.Background(), RegisterInput{
		Email:    "invalid",
		Password: "12345678",
		Nickname: "tester",
	})

	assert.Error(t, err)
	assert.Equal(t, "invalid email format", err.Error())
	assert.Nil(t, user)
	assert.Empty(t, token)
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	service := NewAuthService(mockUserRepo, mockSessionRepo)

	user, token, err := service.Register(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "123",
		Nickname: "tester",
	})

	assert.Error(t, err)
	assert.Equal(t, "password must be at least 8 characters", err.Error())
	assert.Nil(t, user)
	assert.Empty(t, token)
}

func TestAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	service := NewAuthService(mockUserRepo, mockSessionRepo)

	hashedPassword, _ := utils.HashPassword("12345678")
	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		Nickname:     "tester",
		PasswordHash: hashedPassword,
	}

	mockUserRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(user, nil)
	mockSessionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	result, token, err := service.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "12345678",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	service := NewAuthService(mockUserRepo, mockSessionRepo)

	mockUserRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, nil)

	result, token, err := service.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "wrong",
	})

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Nil(t, result)
	assert.Empty(t, token)
}

func TestAuthService_Logout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	service := NewAuthService(mockUserRepo, mockSessionRepo)

	mockSessionRepo.EXPECT().DeleteByToken(gomock.Any(), "token123").Return(nil)

	err := service.Logout(context.Background(), "token123")

	assert.NoError(t, err)
}
