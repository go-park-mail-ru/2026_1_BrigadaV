package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPlaceRepo_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PlaceRepo{db: nil}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "price", "created_at", "updated_at",
		"locality_id", "locality_name", "country_name", "latitude", "longitude",
		"category_id", "category_name", "category_description",
		"photo_id", "photo_place_id", "photo_photo_id", "is_main", "photo_created_at",
		"photo_file_id", "file_path", "file_created_at",
	}).AddRow(1, "Eiffel Tower", "Famous tower", 1500, time.Now(), time.Now(),
		nil, nil, nil, nil, nil,
		nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description,
               ph.id, ph.place_id, ph.photo_id, ph.is_main, ph.created_at,
               pt.id, pt.file_path, pt.created_at
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        LEFT JOIN place_photo ph ON p.id = ph.place_id
        LEFT JOIN photo pt ON ph.photo_id = pt.id
        ORDER BY p.id`)).
		WillReturnRows(rows)

	places, err := repo.GetAll(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, places)
}

func TestPlaceRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PlaceRepo{db: nil}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "price", "created_at", "updated_at",
		"locality_id", "locality_name", "country_name", "latitude", "longitude",
		"category_id", "category_name", "category_description",
	}).AddRow(1, "Eiffel Tower", "Famous tower", 1500, time.Now(), time.Now(),
		1, "Paris", "France", 48.8566, 2.3522,
		2, "Museum", "Art museum")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        WHERE p.id = $1`)).
		WithArgs(1).
		WillReturnRows(rows)

	place, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, place)
	assert.Equal(t, "Eiffel Tower", place.Name)
}

func TestPlaceRepo_GetWithRatingAndLike(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PlaceRepo{db: nil}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM review WHERE place_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"avg", "count"}).AddRow(4.5, 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM favorite WHERE user_id = $1 AND place_id = $2)`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	placeRows := sqlmock.NewRows([]string{
		"id", "name", "description", "price", "created_at", "updated_at",
		"locality_id", "locality_name", "country_name", "latitude", "longitude",
		"category_id", "category_name", "category_description",
	}).AddRow(1, "Eiffel Tower", "Famous tower", 1500, time.Now(), time.Now(),
		1, "Paris", "France", 48.8566, 2.3522,
		2, "Museum", "Art museum")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        WHERE p.id = $1`)).
		WithArgs(1).
		WillReturnRows(placeRows)

	result, err := repo.GetWithRatingAndLike(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 4.5, result.Rating)
	assert.Equal(t, int64(10), result.ReviewCount)
	assert.True(t, result.IsLiked)
}
