package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
)

type ReviewRepo struct {
	db DB
}

func NewReviewRepo(db DB) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (r *ReviewRepo) Create(ctx context.Context, review *models.Review) error {
	query := `INSERT INTO review (user_id, place_id, title, rating, comment, visit_date)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query,
		review.UserID, review.PlaceID, review.Title, review.Rating, review.Comment, review.VisitDate,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
}

func (r *ReviewRepo) GetByID(ctx context.Context, id uint64) (*models.Review, error) {
	query := `SELECT id, user_id, place_id, title, rating, comment, visit_date, created_at, updated_at
              FROM review WHERE id = $1`
	var rev models.Review
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rev.ID, &rev.UserID, &rev.PlaceID, &rev.Title, &rev.Rating, &rev.Comment, &rev.VisitDate,
		&rev.CreatedAt, &rev.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &rev, err
}

func (r *ReviewRepo) GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
	query := `
        SELECT r.id, r.title, r.rating, r.comment, r.created_at,
               u.id, u.nickname, u.avatar_url
        FROM review r
        JOIN "user" u ON r.user_id = u.id
        WHERE r.place_id = $1
        ORDER BY r.created_at DESC
    `
	rows, err := r.db.Query(ctx, query, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reviews []models.ReviewWithAuthor
	for rows.Next() {
		var rev models.ReviewWithAuthor
		var authorID uint64
		var authorNickname string
		var authorAvatar *string
		err := rows.Scan(&rev.ID, &rev.Title, &rev.Rating, &rev.Comment, &rev.CreatedAt,
			&authorID, &authorNickname, &authorAvatar)
		if err != nil {
			return nil, err
		}
		rev.Author.ID = authorID
		rev.Author.Nickname = authorNickname
		rev.Author.Avatar = authorAvatar
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (r *ReviewRepo) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM review WHERE id = $1`, id)
	return err
}
