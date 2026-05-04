package album

import (
	"context"
	"guidely-app/internal/album/repository"
	"guidely-app/pkg/models"
)

type AlbumService interface {
	Create(ctx context.Context, a *models.Album) error
	GetByID(ctx context.Context, id uint64) (*models.Album, error)
	GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error)
	Update(ctx context.Context, a *models.Album) error
	Delete(ctx context.Context, id uint64) error
	UploadPhoto(ctx context.Context, albumID uint64, filePath string) (uint64, error)
	AddPhoto(ctx context.Context, albumID, photoID uint64, order int16) error
	RemovePhoto(ctx context.Context, albumID, photoID uint64) error
	GetPhotos(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error)
}

type albumServiceImpl struct {
	repo repository.AlbumRepository
}

func NewService(repo repository.AlbumRepository) AlbumService {
	return &albumServiceImpl{repo: repo}
}

func (s *albumServiceImpl) Create(ctx context.Context, a *models.Album) error {
	return s.repo.Create(ctx, a)
}

func (s *albumServiceImpl) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *albumServiceImpl) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
	return s.repo.GetByTrip(ctx, tripID)
}

func (s *albumServiceImpl) Update(ctx context.Context, a *models.Album) error {
	return s.repo.Update(ctx, a)
}

func (s *albumServiceImpl) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *albumServiceImpl) UploadPhoto(ctx context.Context, albumID uint64, filePath string) (uint64, error) {
	return s.repo.UploadPhoto(ctx, albumID, filePath)
}

func (s *albumServiceImpl) AddPhoto(ctx context.Context, albumID, photoID uint64, order int16) error {
	return s.repo.AddPhoto(ctx, albumID, photoID, order)
}

func (s *albumServiceImpl) RemovePhoto(ctx context.Context, albumID, photoID uint64) error {
	return s.repo.RemovePhoto(ctx, albumID, photoID)
}

func (s *albumServiceImpl) GetPhotos(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error) {
	return s.repo.GetPhotos(ctx, albumID)
}
