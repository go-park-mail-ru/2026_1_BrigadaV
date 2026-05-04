package repository

import (
	"context"
	"guidely-app/pkg/models"
)

type AlbumRepository interface {
	Create(ctx context.Context, album *models.Album) error
	GetByID(ctx context.Context, id uint64) (*models.Album, error)
	GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error)
	Update(ctx context.Context, album *models.Album) error
	Delete(ctx context.Context, id uint64) error
	// UploadPhoto: сохраняет file_path в таблицу photo, создаёт связь album_photo
	UploadPhoto(ctx context.Context, albumID uint64, filePath string) (uint64, error)
	AddPhoto(ctx context.Context, albumID, photoID uint64, order int16) error
	RemovePhoto(ctx context.Context, albumID, photoID uint64) error
	// GetPhotos возвращает фото с url (file_path из таблицы photo)
	GetPhotos(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error)
}
