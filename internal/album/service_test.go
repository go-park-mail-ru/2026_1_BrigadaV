package album

import (
	"context"
	"errors"
	"testing"

	"guidely-app/internal/album/repository/mocks"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)
	album := &models.Album{TripID: 1, Name: "Test"}
	repo.EXPECT().Create(gomock.Any(), album).Return(nil)
	err := svc.Create(context.Background(), album)
	assert.NoError(t, err)
}

func TestService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)
	expected := &models.Album{ID: 1, Name: "Test"}
	repo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(expected, nil)
	album, err := svc.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, album)
}

func TestService_GetByTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)
	albums := []models.Album{{ID: 1, TripID: 1}, {ID: 2, TripID: 1}}
	repo.EXPECT().GetByTrip(gomock.Any(), uint64(1)).Return(albums, nil)
	result, err := svc.GetByTrip(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)
	album := &models.Album{ID: 1, Name: "Updated"}
	repo.EXPECT().Update(gomock.Any(), album).Return(nil)
	err := svc.Update(context.Background(), album)
	assert.NoError(t, err)
}

func TestService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)
	repo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)
	err := svc.Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestService_GetByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(nil, errors.New("db error"))
	_, err := svc.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestService_GetByTrip_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().GetByTrip(gomock.Any(), uint64(1)).Return(nil, errors.New("db error"))
	_, err := svc.GetByTrip(context.Background(), 1)
	assert.Error(t, err)
}

func TestService_Update_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	album := &models.Album{ID: 1, Name: "Test"}
	repo.EXPECT().Update(gomock.Any(), album).Return(errors.New("db error"))
	err := svc.Update(context.Background(), album)
	assert.Error(t, err)
}

func TestService_Delete_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(errors.New("db error"))
	err := svc.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestService_UploadPhoto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().UploadPhoto(gomock.Any(), uint64(1), "/photos/test.jpg").Return(uint64(10), nil)
	id, err := svc.UploadPhoto(context.Background(), 1, "/photos/test.jpg")
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), id)
}

func TestService_UploadPhoto_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().UploadPhoto(gomock.Any(), uint64(1), "/photos/test.jpg").Return(uint64(0), errors.New("db error"))
	_, err := svc.UploadPhoto(context.Background(), 1, "/photos/test.jpg")
	assert.Error(t, err)
}

func TestService_AddPhoto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().AddPhoto(gomock.Any(), uint64(1), uint64(5), int16(2)).Return(nil)
	err := svc.AddPhoto(context.Background(), 1, 5, 2)
	assert.NoError(t, err)
}

func TestService_RemovePhoto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().RemovePhoto(gomock.Any(), uint64(1), uint64(5)).Return(nil)
	err := svc.RemovePhoto(context.Background(), 1, 5)
	assert.NoError(t, err)
}

func TestService_GetPhotos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	expected := []models.AlbumPhoto{{AlbumID: 1, PhotoID: 10, OrderIndex: 1}}
	repo.EXPECT().GetPhotos(gomock.Any(), uint64(1)).Return(expected, nil)
	photos, err := svc.GetPhotos(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, photos, 1)
}
