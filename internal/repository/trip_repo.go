package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepo struct {
	db *pgxpool.Pool
}

func NewTripRepo(db *pgxpool.Pool) *TripRepo {
	return &TripRepo{db: db}
}

func (r *TripRepo) Create(ctx context.Context, trip *models.Trip) error {
	query := `INSERT INTO trip (title, description, start_date, end_date, created_by, is_public)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, trip.Title, trip.Description, trip.StartDate, trip.EndDate, trip.CreatedBy, trip.IsPublic).
		Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)
	return err
}

func (r *TripRepo) GetByID(ctx context.Context, id uint64) (*models.Trip, error) {
	query := `SELECT id, title, description, start_date, end_date, created_by, is_public, created_at, updated_at
              FROM trip WHERE id = $1`
	var trip models.Trip
	err := r.db.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.Title, &trip.Description, &trip.StartDate, &trip.EndDate,
		&trip.CreatedBy, &trip.IsPublic, &trip.CreatedAt, &trip.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &trip, err
}

func (r *TripRepo) GetByUser(ctx context.Context, userID uint64) ([]models.Trip, error) {
	query := `SELECT id, title, description, start_date, end_date, created_by, is_public, created_at, updated_at
              FROM trip WHERE created_by = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var trips []models.Trip
	for rows.Next() {
		var t models.Trip
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.StartDate, &t.EndDate,
			&t.CreatedBy, &t.IsPublic, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		trips = append(trips, t)
	}
	return trips, nil
}

func (r *TripRepo) Update(ctx context.Context, trip *models.Trip) error {
	query := `UPDATE trip SET title = $1, description = $2, start_date = $3, end_date = $4, is_public = $5, updated_at = NOW()
              WHERE id = $6 RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, trip.Title, trip.Description, trip.StartDate, trip.EndDate, trip.IsPublic, trip.ID).
		Scan(&trip.UpdatedAt)
	return err
}

func (r *TripRepo) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM trip WHERE id = $1`, id)
	return err
}
