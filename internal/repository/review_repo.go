package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepo struct {
	db *pgxpool.Pool
}

func NewReviewRepo(db *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (r *ReviewRepo) Create(ctx context.Context, review *models.Review) error {
	query := `INSERT INTO review (user_id, place_id, rating, comment, visit_date)
              VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, review.UserID, review.PlaceID, review.Rating, review.Comment, review.VisitDate).
		Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
	return err
}

func (r *ReviewRepo) GetByID(ctx context.Context, id uint64) (*models.Review, error) {
	query := `SELECT id, user_id, place_id, rating, comment, visit_date, created_at, updated_at
              FROM review WHERE id = $1`
	var rev models.Review
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rev.ID, &rev.UserID, &rev.PlaceID, &rev.Rating, &rev.Comment, &rev.VisitDate,
		&rev.CreatedAt, &rev.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &rev, err
}

func (r *ReviewRepo) GetByPlaceID(ctx context.Context, placeID uint64) ([]models.Review, error) {
	query := `SELECT id, user_id, place_id, rating, comment, visit_date, created_at, updated_at
              FROM review WHERE place_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reviews []models.Review
	for rows.Next() {
		var rev models.Review
		err := rows.Scan(&rev.ID, &rev.UserID, &rev.PlaceID, &rev.Rating, &rev.Comment, &rev.VisitDate,
			&rev.CreatedAt, &rev.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (r *ReviewRepo) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM review WHERE id = $1`, id)
	return err
}
