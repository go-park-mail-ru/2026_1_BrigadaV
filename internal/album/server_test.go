package album

import (
	"context"
	"errors"
	"testing"

	"guidely-app/pkg/models"
	pb "guidely-app/pkg/pb/album"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockAlbumService struct {
	createFn      func(ctx context.Context, a *models.Album) error
	getByIDFn     func(ctx context.Context, id uint64) (*models.Album, error)
	getByTripFn   func(ctx context.Context, tripID uint64) ([]models.Album, error)
	updateFn      func(ctx context.Context, a *models.Album) error
	deleteFn      func(ctx context.Context, id uint64) error
	addPhotoFn    func(ctx context.Context, albumID, photoID uint64, order int16) error
	removePhotoFn func(ctx context.Context, albumID, photoID uint64) error
	getPhotosFn   func(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error)
	uploadPhotoFn func(ctx context.Context, albumID uint64, filePath string) (uint64, error)
}

func (m *mockAlbumService) Create(ctx context.Context, a *models.Album) error {
	return m.createFn(ctx, a)
}
func (m *mockAlbumService) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockAlbumService) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
	return m.getByTripFn(ctx, tripID)
}
func (m *mockAlbumService) Update(ctx context.Context, a *models.Album) error {
	return m.updateFn(ctx, a)
}
func (m *mockAlbumService) Delete(ctx context.Context, id uint64) error { return m.deleteFn(ctx, id) }
func (m *mockAlbumService) AddPhoto(ctx context.Context, albumID, photoID uint64, order int16) error {
	return m.addPhotoFn(ctx, albumID, photoID, order)
}
func (m *mockAlbumService) RemovePhoto(ctx context.Context, albumID, photoID uint64) error {
	return m.removePhotoFn(ctx, albumID, photoID)
}
func (m *mockAlbumService) GetPhotos(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error) {
	return m.getPhotosFn(ctx, albumID)
}
func (m *mockAlbumService) UploadPhoto(ctx context.Context, albumID uint64, filePath string) (uint64, error) {
	return m.uploadPhotoFn(ctx, albumID, filePath)
}

func TestServer_Create(t *testing.T) {
	svc := &mockAlbumService{createFn: func(ctx context.Context, a *models.Album) error { return nil }}
	srv := NewServer(svc)
	_, err := srv.Create(context.Background(), &pb.CreateAlbumRequest{TripId: 1, Name: "Test"})
	assert.NoError(t, err)
}

func TestServer_Create_Error(t *testing.T) {
	svc := &mockAlbumService{createFn: func(ctx context.Context, a *models.Album) error { return errors.New("db error") }}
	srv := NewServer(svc)
	_, err := srv.Create(context.Background(), &pb.CreateAlbumRequest{TripId: 1, Name: "Test"})
	assert.Error(t, err)
}

func TestServer_Get(t *testing.T) {
	svc := &mockAlbumService{getByIDFn: func(ctx context.Context, id uint64) (*models.Album, error) {
		return &models.Album{ID: 1, Name: "Test"}, nil
	}}
	srv := NewServer(svc)
	resp, err := srv.Get(context.Background(), &pb.GetAlbumRequest{Id: 1})
	assert.NoError(t, err)
	assert.Equal(t, "Test", resp.Name)
}

func TestServer_Get_NotFound(t *testing.T) {
	svc := &mockAlbumService{getByIDFn: func(ctx context.Context, id uint64) (*models.Album, error) { return nil, errors.New("not found") }}
	srv := NewServer(svc)
	_, err := srv.Get(context.Background(), &pb.GetAlbumRequest{Id: 1})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestServer_GetByTrip(t *testing.T) {
	svc := &mockAlbumService{getByTripFn: func(ctx context.Context, tripID uint64) ([]models.Album, error) {
		return []models.Album{{ID: 1}, {ID: 2}}, nil
	}}
	srv := NewServer(svc)
	resp, err := srv.GetByTrip(context.Background(), &pb.GetAlbumByTripRequest{TripId: 1})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestServer_Update(t *testing.T) {
	svc := &mockAlbumService{updateFn: func(ctx context.Context, a *models.Album) error { return nil }}
	srv := NewServer(svc)
	_, err := srv.Update(context.Background(), &pb.UpdateAlbumRequest{Id: 1, Name: "Updated"})
	assert.NoError(t, err)
}

func TestServer_Update_Error(t *testing.T) {
	svc := &mockAlbumService{updateFn: func(ctx context.Context, a *models.Album) error { return errors.New("db error") }}
	srv := NewServer(svc)
	_, err := srv.Update(context.Background(), &pb.UpdateAlbumRequest{Id: 1, Name: "Updated"})
	assert.Error(t, err)
}

func TestServer_Delete(t *testing.T) {
	svc := &mockAlbumService{deleteFn: func(ctx context.Context, id uint64) error { return nil }}
	srv := NewServer(svc)
	_, err := srv.Delete(context.Background(), &pb.DeleteAlbumRequest{Id: 1})
	assert.NoError(t, err)
}

func TestServer_Delete_Error(t *testing.T) {
	svc := &mockAlbumService{deleteFn: func(ctx context.Context, id uint64) error { return errors.New("db error") }}
	srv := NewServer(svc)
	_, err := srv.Delete(context.Background(), &pb.DeleteAlbumRequest{Id: 1})
	assert.Error(t, err)
}

func TestServer_AddPhoto(t *testing.T) {
	svc := &mockAlbumService{addPhotoFn: func(ctx context.Context, albumID, photoID uint64, order int16) error { return nil }}
	srv := NewServer(svc)
	resp, err := srv.AddPhoto(context.Background(), &pb.AddPhotoRequest{AlbumId: 1, PhotoId: 10, Order: 1})
	assert.NoError(t, err)
	assert.Equal(t, int32(1), resp.OrderIndex)
}

func TestServer_RemovePhoto(t *testing.T) {
	svc := &mockAlbumService{removePhotoFn: func(ctx context.Context, albumID, photoID uint64) error { return nil }}
	srv := NewServer(svc)
	_, err := srv.RemovePhoto(context.Background(), &pb.RemovePhotoRequest{AlbumId: 1, PhotoId: 10})
	assert.NoError(t, err)
}

func TestServer_GetPhotos(t *testing.T) {
	svc := &mockAlbumService{getPhotosFn: func(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error) {
		return []models.AlbumPhoto{{AlbumID: 1, PhotoID: 10, OrderIndex: 1}}, nil
	}}
	srv := NewServer(svc)
	resp, err := srv.GetPhotos(context.Background(), &pb.GetAlbumPhotosRequest{AlbumId: 1})
	assert.NoError(t, err)
	assert.Len(t, resp.Photos, 1)
}
