package repository

import (
	"context"
	"testing"
	"time"

	"guidely-app/internal/models"
	"guidely-app/internal/testutil"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestTripRepo_Create(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripRepo(mockPool)

	trip := &models.Trip{
		Title:      "My Trip",
		Location:   testutil.PtrString("Paris"),
		StartDate:  nil,
		EndDate:    nil,
		PreviewURL: testutil.PtrString("/preview.jpg"),
		CreatedBy:  1,
		IsPublic:   true,
	}

	rows := mockPool.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(1, time.Now(), time.Now())

	mockPool.ExpectQuery(`INSERT INTO trip \(title, description, location, start_date, end_date, preview_url, created_by, is_public\)`).
		WithArgs(trip.Title, trip.Description, trip.Location, trip.StartDate, trip.EndDate, trip.PreviewURL, trip.CreatedBy, trip.IsPublic).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), trip)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), trip.ID)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripRepo_GetByID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripRepo(mockPool)

	rows := mockPool.NewRows([]string{"id", "title", "description", "location", "start_date", "end_date", "preview_url", "created_by", "is_public", "created_at", "updated_at"}).
		AddRow(1, "My Trip", nil, "Paris", nil, nil, "/preview.jpg", 1, true, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, title, description, location, start_date, end_date, preview_url, created_by, is_public, created_at, updated_at FROM trip WHERE id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	trip, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, "My Trip", trip.Title)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripRepo_GetByUser(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripRepo(mockPool)

	rows := mockPool.NewRows([]string{"id", "title", "description", "location", "start_date", "end_date", "preview_url", "created_by", "is_public", "created_at", "updated_at"}).
		AddRow(1, "Trip1", nil, "Paris", nil, nil, "/preview1.jpg", 1, true, time.Now(), time.Now()).
		AddRow(2, "Trip2", nil, "London", nil, nil, "/preview2.jpg", 1, false, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, title, description, location, start_date, end_date, preview_url, created_by, is_public, created_at, updated_at FROM trip WHERE created_by = \$1 ORDER BY created_at DESC`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	trips, err := repo.GetByUser(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, trips, 2)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripRepo_Update(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripRepo(mockPool)

	trip := &models.Trip{
		ID:          1,
		Title:       "Updated Trip",
		Description: "New description",
		Location:    testutil.PtrString("Paris"),
		IsPublic:    false,
	}

	rows := mockPool.NewRows([]string{"updated_at"}).AddRow(time.Now())

	mockPool.ExpectQuery(`UPDATE trip SET title = \$1, description = \$2, location = \$3, start_date = \$4, end_date = \$5, preview_url = \$6, is_public = \$7, updated_at = NOW\(\) WHERE id = \$8 RETURNING updated_at`).
		WithArgs(trip.Title, trip.Description, trip.Location, trip.StartDate, trip.EndDate, trip.PreviewURL, trip.IsPublic, trip.ID).
		WillReturnRows(rows)

	err = repo.Update(context.Background(), trip)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripRepo_Delete(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM trip WHERE id = \$1`).
		WithArgs(uint64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripRepo_AddAttraction(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripRepo(mockPool)

	mockPool.ExpectExec(`INSERT INTO trip_attractions \(trip_id, place_id, order_index\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(uint64(1), uint64(1), int16(0)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.AddAttraction(context.Background(), 1, 1, 0)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripRepo_GetAttractions(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripRepo(mockPool)

	rows := mockPool.NewRows([]string{"id", "name", "description", "rating", "image_photo_id"}).
		AddRow(1, "Eiffel Tower", "Famous tower", 4.5, nil)

	// В запросе есть COALESCE(AVG(r.rating),0) - pgxmock проверяет только начало запроса
	mockPool.ExpectQuery(`SELECT p\.id, p\.name, p\.description, COALESCE\(AVG\(r\.rating\), 0\) as rating,`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	places, err := repo.GetAttractions(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, places, 1)
	assert.Equal(t, "Eiffel Tower", places[0].Name)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
