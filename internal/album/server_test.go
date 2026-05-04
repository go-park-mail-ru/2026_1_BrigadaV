package album

// import (
// 	"context"
// 	"testing"

// 	"guidely-app/pkg/models"
// 	pb "guidely-app/pkg/pb/album"

// 	"github.com/stretchr/testify/assert"
// )

// type mockAlbumService struct{}

// func (m *mockAlbumService) Create(ctx context.Context, a *models.Album) error { return nil }
// func (m *mockAlbumService) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
// 	return nil, nil
// }
// func (m *mockAlbumService) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
// 	return nil, nil
// }
// func (m *mockAlbumService) Update(ctx context.Context, a *models.Album) error { return nil }
// func (m *mockAlbumService) Delete(ctx context.Context, id uint64) error       { return nil }
// func (m *mockAlbumService) AddPhoto(ctx context.Context, albumID, photoID uint64, order int16) error {
// 	return nil
// }
// func (m *mockAlbumService) RemovePhoto(ctx context.Context, albumID, photoID uint64) error {
// 	return nil
// }
// func (m *mockAlbumService) GetPhotos(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error) {
// 	return nil, nil
// }

// func TestServer_Create(t *testing.T) {
// 	svc := &mockAlbumService{}
// 	srv := NewServer(svc)
// 	_, err := srv.Create(context.Background(), &pb.CreateAlbumRequest{
// 		TripId: 1, Name: "Test",
// 	})
// 	assert.NoError(t, err)
// }
