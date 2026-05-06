package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"guidely-app/pkg/models"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestAlbumRepo_Create(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	album := &models.Album{
		TripID:      1,
		Name:        "Test Album",
		Description: "desc",
		MaxPhotos:   50,
	}

	rows := mockPool.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(uint64(1), time.Now(), time.Now())

	mockPool.ExpectQuery(`INSERT INTO album`).
		WithArgs(album.TripID, album.Name, album.Description, album.CoverPhotoID, album.MaxPhotos).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), album)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), album.ID)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_GetByID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	rows := mockPool.NewRows([]string{
		"id", "trip_id", "name", "description", "cover_photo_id", "max_photos", "created_at", "updated_at",
	}).AddRow(uint64(1), uint64(1), "Test", "desc", nil, 50, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, trip_id, name, description, cover_photo_id, max_photos, created_at, updated_at FROM album WHERE id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	album, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, album)
	assert.Equal(t, "Test", album.Name)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_GetByTrip(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	rows := mockPool.NewRows([]string{
		"id", "trip_id", "name", "description", "cover_photo_id", "max_photos", "created_at", "updated_at",
	}).AddRow(uint64(1), uint64(1), "Album1", "", nil, 10, time.Now(), time.Now()).
		AddRow(uint64(2), uint64(1), "Album2", "", nil, 20, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, trip_id, name, description, cover_photo_id, max_photos, created_at, updated_at FROM album WHERE trip_id = \$1 ORDER BY created_at`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	albums, err := repo.GetByTrip(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, albums, 2)
	assert.Equal(t, "Album1", albums[0].Name)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_Update(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	album := &models.Album{
		ID:           1,
		Name:         "Updated",
		Description:  "new desc",
		CoverPhotoID: ptrUint64(5),
		MaxPhotos:    30,
	}

	mockPool.ExpectExec(`UPDATE album SET name=\$1, description=\$2, cover_photo_id=\$3, max_photos=\$4, updated_at=NOW\(\) WHERE id=\$5`).
		WithArgs(album.Name, album.Description, album.CoverPhotoID, album.MaxPhotos, album.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Update(context.Background(), album)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_Delete(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM album WHERE id=\$1`).
		WithArgs(uint64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), 1)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_AddPhoto(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	mockPool.ExpectExec(`INSERT INTO album_photo \(album_id, photo_id, order_index, created_at\)`).
		WithArgs(uint64(1), uint64(10), int16(0)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.AddPhoto(context.Background(), 1, 10, 0)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_RemovePhoto(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM album_photo WHERE album_id=\$1 AND photo_id=\$2`).
		WithArgs(uint64(1), uint64(10)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.RemovePhoto(context.Background(), 1, 10)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_GetPhotos(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	rows := mockPool.NewRows([]string{"album_id", "photo_id", "order_index", "created_at", "file_path"}).
		AddRow(uint64(1), uint64(10), int16(1), time.Now(), "/photos/img.jpg")

	mockPool.ExpectQuery(`SELECT ap\.album_id, ap\.photo_id, ap\.order_index, ap\.created_at, p\.file_path`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	photos, err := repo.GetPhotos(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, photos, 1)
	assert.Equal(t, "/photos/img.jpg", photos[0].FilePath)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_UploadPhoto(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewAlbumRepo(mockPool)

	mockPool.ExpectBegin()
	mockPool.ExpectQuery(`INSERT INTO photo \(file_path, created_at\)`).
		WithArgs("/photos/test.jpg").
		WillReturnRows(mockPool.NewRows([]string{"id"}).AddRow(uint64(100)))
	mockPool.ExpectQuery(`SELECT COALESCE\(MAX\(order_index\), 0\) FROM album_photo WHERE album_id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(mockPool.NewRows([]string{"max"}).AddRow(int64(2)))
	mockPool.ExpectExec(`INSERT INTO album_photo \(album_id, photo_id, order_index, created_at\)`).
		WithArgs(uint64(1), uint64(100), 3). // int
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockPool.ExpectCommit()

	photoID, err := repo.UploadPhoto(context.Background(), 1, "/photos/test.jpg")
	assert.NoError(t, err)
	assert.Equal(t, uint64(100), photoID)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestAlbumRepo_Create_DBError(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewAlbumRepo(mockPool)

	album := &models.Album{TripID: 1, Name: "Test"}
	mockPool.ExpectQuery(`INSERT INTO album`).WillReturnError(errors.New("db down"))
	err := repo.Create(context.Background(), album)
	assert.Error(t, err)
}

func TestAlbumRepo_GetByID_NotFound(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewAlbumRepo(mockPool)

	mockPool.ExpectQuery(`SELECT .+ FROM album WHERE id = \$1`).WithArgs(uint64(1)).WillReturnRows(mockPool.NewRows(nil))
	album, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Nil(t, album)
}

func TestAlbumRepo_GetByTrip_Empty(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewAlbumRepo(mockPool)

	mockPool.ExpectQuery(`SELECT .+ FROM album WHERE trip_id = \$1`).WithArgs(uint64(1)).WillReturnRows(mockPool.NewRows(nil))
	albums, err := repo.GetByTrip(context.Background(), 1)
	assert.NoError(t, err)
	assert.Empty(t, albums)
}

func ptrUint64(v uint64) *uint64 { return &v }
