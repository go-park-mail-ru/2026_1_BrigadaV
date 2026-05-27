package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestCountryRepo_GetAll_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	repo := NewCountryRepo(mockPool)

	rows := mockPool.NewRows([]string{"id", "name", "created_at"}).
		AddRow(uint64(1), "France", time.Now()).
		AddRow(uint64(2), "Italy", time.Now())
	mockPool.ExpectQuery(`SELECT id, name, created_at FROM country ORDER BY name`).
		WillReturnRows(rows)

	countries, err := repo.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, countries, 2)
}

func TestCountryRepo_GetAll_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	repo := NewCountryRepo(mockPool)

	mockPool.ExpectQuery(`SELECT id, name, created_at FROM country ORDER BY name`).
		WillReturnError(errors.New("db error"))
	_, err = repo.GetAll(context.Background())
	assert.Error(t, err)
}

func TestCountryRepo_GetLocalitiesByCountryID_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	repo := NewCountryRepo(mockPool)

	lat := 48.8566
	lng := 2.3522
	rows := mockPool.NewRows([]string{"id", "name", "latitude", "longitude"}).
		AddRow(uint64(10), "Paris", &lat, &lng) // передаём указатели
	mockPool.ExpectQuery(`SELECT id, name, latitude, longitude FROM locality WHERE country_id = \$1 ORDER BY name`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	loc, err := repo.GetLocalitiesByCountryID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, loc, 1)
}
