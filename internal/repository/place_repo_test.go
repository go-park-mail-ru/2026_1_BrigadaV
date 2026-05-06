package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestPlaceRepo_GetAll(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewPlaceRepo(mockPool)

	localityName := "Paris"
	countryName := "France"
	latitude := 48.8566
	longitude := 2.3522
	categoryName := "Museum"
	categoryDesc := "Art museum"
	photoFilePath := "/photos/eiffel.jpg"
	placePhotoID := uint64(1)
	isMain := true

	rows := mockPool.NewRows([]string{
		"id", "name", "description", "photo_url", "price", "created_at", "updated_at",
		"place_lat", "place_lng", // p.latitude, p.longitude
		"locality_id", "locality_name", "country_name", "loc_lat", "loc_lng",
		"category_id", "category_name", "category_description",
		"place_photo_id", "file_path", "is_main",
	}).AddRow(
		uint64(1), "Eiffel Tower", "Famous tower", nil, 1500, time.Now(), time.Now(),
		&latitude, &longitude, // указатели, т.к. Scan ожидает *float64
		nil, &localityName, &countryName, &latitude, &longitude,
		nil, &categoryName, &categoryDesc,
		&placePhotoID, &photoFilePath, &isMain,
	)

	mockPool.ExpectQuery(`SELECT p\.id, p\.name, p\.description, p\.photo_url, p\.price, p\.created_at, p\.updated_at,`).
		WillReturnRows(rows)

	places, err := repo.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, places, 1)
	assert.Equal(t, "Eiffel Tower", places[0].Name)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPlaceRepo_GetByID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewPlaceRepo(mockPool)

	localityName := "Paris"
	countryName := "France"
	latitude := 48.8566
	longitude := 2.3522
	categoryName := "Museum"
	categoryDesc := "Art museum"

	// теперь добавляем place_lat и place_lng (p.latitude, p.longitude)
	rows := mockPool.NewRows([]string{
		"id", "name", "description", "photo_url", "price", "created_at", "updated_at",
		"place_lat", "place_lng", // новое
		"locality_id", "locality_name", "country_name", "loc_lat", "loc_lng",
		"category_id", "category_name", "category_description",
	}).AddRow(uint64(1), "Eiffel Tower", "Famous tower", nil, 1500, time.Now(), time.Now(),
		&latitude, &longitude, // указатели
		nil, &localityName, &countryName, &latitude, &longitude,
		nil, &categoryName, &categoryDesc)

	mockPool.ExpectQuery(`SELECT p\.id, p\.name, p\.description, p\.photo_url, p\.price, p\.created_at, p\.updated_at,`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	place, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, place)
	assert.Equal(t, "Eiffel Tower", place.Name)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPlaceRepo_GetWithRatingAndLike(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewPlaceRepo(mockPool)

	lat := 48.8566
	lng := 2.3522

	// Запрос теперь возвращает 9 колонок (добавлены latitude, longitude)
	placeRows := mockPool.NewRows([]string{
		"id", "name", "description", "photo_url", "price", "rating", "review_count",
		"latitude", "longitude",
	}).AddRow(uint64(1), "Eiffel Tower", "Famous tower", nil, 1500, 4.5, int64(10),
		&lat, &lng) // указатели

	mockPool.ExpectQuery(`SELECT id, name, description, photo_url, price, rating, review_count, latitude, longitude FROM place WHERE id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(placeRows)

	// Проверка лайка
	likeRows := mockPool.NewRows([]string{"exists"}).AddRow(true)
	mockPool.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM favorite WHERE user_id=\$1 AND place_id=\$2\)`).
		WithArgs(uint64(1), uint64(1)).
		WillReturnRows(likeRows)

	result, err := repo.GetWithRatingAndLike(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 4.5, result.Rating)
	assert.Equal(t, int64(10), result.ReviewCount)
	assert.True(t, result.IsLiked)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
