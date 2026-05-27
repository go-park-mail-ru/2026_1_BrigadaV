package repository

import (
	"context"
	"errors"
	"guidely-app/pkg/models"
	"testing"
	"time"

	"github.com/lib/pq"
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

func TestCategoryRepo_GetByID(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewCategoryRepo(mockPool)

	rows := mockPool.NewRows([]string{"id", "name", "description", "applicable_types", "created_at"}).
		AddRow(uint64(1), "Hotel", "desc", []string{"hotel"}, time.Now())
	mockPool.ExpectQuery(`SELECT id, name, description, applicable_types, created_at FROM category WHERE id = \$1`).WithArgs(uint64(1)).WillReturnRows(rows)
	cat, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, cat)
	assert.Equal(t, "Hotel", cat.Name)
}

func TestCategoryRepo_Create_DBError(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewCategoryRepo(mockPool)

	cat := &models.Category{Name: "Test", ApplicableTypes: []string{"attraction"}}
	mockPool.ExpectQuery(`INSERT INTO category`).WillReturnError(errors.New("db error"))
	err := repo.Create(context.Background(), cat)
	assert.Error(t, err)
}

func TestCategoryRepo_Update_DBError(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewCategoryRepo(mockPool)

	cat := &models.Category{ID: 1, Name: "Updated", ApplicableTypes: []string{"hotel"}}
	mockPool.ExpectExec(`UPDATE category`).WithArgs(cat.Name, cat.Description, pq.Array(cat.ApplicableTypes), cat.ID).WillReturnError(errors.New("db error"))
	err := repo.Update(context.Background(), cat)
	assert.Error(t, err)
}

func TestCategoryRepo_Delete_DBError(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewCategoryRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM category WHERE id=\$1`).WithArgs(uint64(1)).WillReturnError(errors.New("db error"))
	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}
