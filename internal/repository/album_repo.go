package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AlbumRepo struct {
	db *pgxpool.Pool
}

func NewAlbumRepo(db *pgxpool.Pool) *AlbumRepo {
	return &AlbumRepo{db: db}
}

func (r *AlbumRepo) Create(ctx context.Context, album *models.Album) error {
	query := `INSERT INTO album (trip_id, name, description, cover_photo_id, max_photos, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	          RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query,
		album.TripID, album.Name, album.Description, album.CoverPhotoID, album.MaxPhotos,
	).Scan(&album.ID, &album.CreatedAt, &album.UpdatedAt)
}

func (r *AlbumRepo) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
	query := `SELECT id, trip_id, name, description, cover_photo_id, max_photos, created_at, updated_at
	          FROM album WHERE id = $1`
	var album models.Album
	err := r.db.QueryRow(ctx, query, id).Scan(
		&album.ID, &album.TripID, &album.Name, &album.Description,
		&album.CoverPhotoID, &album.MaxPhotos, &album.CreatedAt, &album.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &album, err
}

func (r *AlbumRepo) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
	query := `SELECT id, trip_id, name, description, cover_photo_id, max_photos, created_at, updated_at
	          FROM album WHERE trip_id = $1 ORDER BY created_at`
	rows, err := r.db.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var albums []models.Album
	for rows.Next() {
		var a models.Album
		if err := rows.Scan(&a.ID, &a.TripID, &a.Name, &a.Description,
			&a.CoverPhotoID, &a.MaxPhotos, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}
	return albums, nil
}

func (r *AlbumRepo) Update(ctx context.Context, album *models.Album) error {
	query := `UPDATE album SET name=$1, description=$2, cover_photo_id=$3, max_photos=$4, updated_at=NOW()
	          WHERE id=$5`
	_, err := r.db.Exec(ctx, query, album.Name, album.Description, album.CoverPhotoID, album.MaxPhotos, album.ID)
	return err
}

func (r *AlbumRepo) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM album WHERE id=$1`, id)
	return err
}

func (r *AlbumRepo) AddPhoto(ctx context.Context, albumID, photoID uint64, order int16) error {
	_, err := r.db.Exec(ctx, `INSERT INTO album_photo (album_id, photo_id, order_index, created_at)
		VALUES ($1, $2, $3, NOW())`, albumID, photoID, order)
	return err
}

func (r *AlbumRepo) RemovePhoto(ctx context.Context, albumID, photoID uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM album_photo WHERE album_id=$1 AND photo_id=$2`, albumID, photoID)
	return err
}

func (r *AlbumRepo) GetPhotos(ctx context.Context, albumID uint64) ([]models.AlbumPhoto, error) {
	query := `SELECT album_id, photo_id, order_index, created_at
	          FROM album_photo WHERE album_id = $1 ORDER BY order_index`
	rows, err := r.db.Query(ctx, query, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var photos []models.AlbumPhoto
	for rows.Next() {
		var p models.AlbumPhoto
		if err := rows.Scan(&p.AlbumID, &p.PhotoID, &p.OrderIndex, &p.CreatedAt); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, nil
}
