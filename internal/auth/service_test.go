package auth

import (
	"context"
	"testing"

	"guidely-app/internal/auth/repository/mocks"
	"guidely-app/pkg/models"
	"guidely-app/pkg/utils"

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
	user, token, err := svc.Register(context.Background(), RegisterInput{
		Login: "test@example.com", Password: "12345678", Nickname: "tester",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, uint64(1), user.ID)
}

func TestService_Register_InvalidLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	_, _, err := svc.Register(context.Background(), RegisterInput{
		Login: "invalid", Password: "12345678", Nickname: "tester",
	})
	assert.Error(t, err)
	assert.Equal(t, "invalid login format", err.Error())
}

func TestService_Register_ShortPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	_, _, err := svc.Register(context.Background(), RegisterInput{
		Login: "test@example.com", Password: "123", Nickname: "tester",
	})
	assert.Error(t, err)
	assert.Equal(t, "password too short", err.Error())
}

func TestService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	hashed, _ := utils.HashPassword("12345678")
	user := &models.User{ID: 1, Login: "test@example.com", PasswordHash: hashed}
	userRepo.EXPECT().GetByLogin(gomock.Any(), "test@example.com").Return(user, nil)
	sessRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	_, token, err := svc.Login(context.Background(), LoginInput{Login: "test@example.com", Password: "12345678"})
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestService_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	hashed, _ := utils.HashPassword("correct")
	user := &models.User{ID: 1, Login: "test@example.com", PasswordHash: hashed}
	userRepo.EXPECT().GetByLogin(gomock.Any(), "test@example.com").Return(user, nil)

	_, _, err := svc.Login(context.Background(), LoginInput{Login: "test@example.com", Password: "wrong"})
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestService_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	token := "test_token"
	hashed := utils.HashToken(token)
	sessRepo.EXPECT().DeleteByToken(gomock.Any(), hashed).Return(nil)

	err := svc.Logout(context.Background(), token)
	assert.NoError(t, err)
}

func TestService_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	expected := &models.User{ID: 1, Nickname: "test"}
	userRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(expected, nil)

	user, err := svc.GetUserByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, user)
}

func TestService_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	user := &models.User{ID: 1, Nickname: "old", AvatarURL: "/old.jpg"}
	newNick := "new"
	newAvatar := "/new.jpg"

	userRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(user, nil)
	userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	updated, err := svc.UpdateProfile(context.Background(), 1, &newNick, &newAvatar, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "new", updated.Nickname)
	assert.Equal(t, "/new.jpg", updated.AvatarURL)
}

func TestService_UpdateAvatar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := mocks.NewMockUserRepository(ctrl)
	sessRepo := mocks.NewMockSessionRepository(ctrl)
	svc := NewService(userRepo, sessRepo)

	user := &models.User{ID: 1, AvatarURL: "/old.jpg"}
	userRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(user, nil)
	userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	updated, err := svc.UpdateAvatar(context.Background(), 1, "/new.jpg")
	assert.NoError(t, err)
	assert.Equal(t, "/new.jpg", updated.AvatarURL)
}
