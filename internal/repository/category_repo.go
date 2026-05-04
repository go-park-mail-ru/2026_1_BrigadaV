package repository

import (
	"context"
	"errors"
	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

type CategoryRepo struct {
	db DB
}

func NewCategoryRepo(db DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) GetAll(ctx context.Context) ([]models.Category, error) {
	query := `SELECT id, name, description, applicable_types, created_at FROM category ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, pq.Array(&c.ApplicableTypes), &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uint64) (*models.Category, error) {
	query := `SELECT id, name, description, applicable_types, created_at FROM category WHERE id = $1`
	var c models.Category
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.Description, pq.Array(&c.ApplicableTypes), &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepo) Create(ctx context.Context, c *models.Category) error {
	query := `INSERT INTO category (name, description, applicable_types) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, c.Name, c.Description, pq.Array(c.ApplicableTypes)).Scan(&c.ID, &c.CreatedAt)
}

func (r *CategoryRepo) Update(ctx context.Context, c *models.Category) error {
	query := `UPDATE category SET name=$1, description=$2, applicable_types=$3 WHERE id=$4`
	_, err := r.db.Exec(ctx, query, c.Name, c.Description, pq.Array(c.ApplicableTypes), c.ID)
	return err
}

func (r *CategoryRepo) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM category WHERE id=$1`, id)
	return err
}
