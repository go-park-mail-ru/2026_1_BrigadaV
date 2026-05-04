package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestCategoryRepo_GetAll(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewCategoryRepo(mockPool)

	rows := mockPool.NewRows([]string{"id", "name", "description", "applicable_types", "created_at"}).
		AddRow(uint64(1), "Отель", "Гостиницы", []string{"hotel"}, time.Now())

	mockPool.ExpectQuery(`SELECT id, name, description, applicable_types, created_at FROM category ORDER BY id`).
		WillReturnRows(rows)

	categories, err := repo.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, categories, 1)
	assert.Equal(t, "Отель", categories[0].Name)
}
