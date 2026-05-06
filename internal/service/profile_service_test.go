package service

import (
	"context"
	"testing"

	"guidely-app/internal/auth/repository/mocks" // моки для UserRepository теперь здесь
	"guidely-app/internal/testutil"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestProfileService_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewProfileService(mockUserRepo)

	expectedUser := &models.User{ID: 1, Nickname: "johnny"}
	mockUserRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(expectedUser, nil)

	user, err := svc.GetProfile(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "johnny", user.Nickname)
}

func TestProfileService_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewProfileService(mockUserRepo)

	existingUser := &models.User{ID: 1, Nickname: "old", AvatarURL: "/old.jpg"}
	mockUserRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(existingUser, nil)
	mockUserRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *models.User) error {
		assert.Equal(t, "new", u.Nickname)
		assert.Equal(t, "/new.jpg", u.AvatarURL)
		return nil
	})

	input := UpdateProfileInput{
		Nickname:  testutil.PtrString("new"),
		AvatarURL: testutil.PtrString("/new.jpg"),
	}
	user, err := svc.UpdateProfile(context.Background(), 1, input)
	assert.NoError(t, err)
	assert.Equal(t, "new", user.Nickname)
}

func TestProfileService_UpdateAvatar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockUserRepository(ctrl)
	svc := NewProfileService(repo)

	user := &models.User{ID: 1, AvatarURL: "/old.jpg"}
	repo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(user, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, u *models.User) error {
		assert.Equal(t, "/new.jpg", u.AvatarURL)
		return nil
	})

	result, err := svc.UpdateAvatar(context.Background(), 1, "/new.jpg")
	assert.NoError(t, err)
	assert.Equal(t, "/new.jpg", result.AvatarURL)
}
