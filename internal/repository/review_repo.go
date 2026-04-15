package repository

import (
	"context"
	"errors"
	"guidely-app/internal/logger"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type ReviewRepo struct {
	db *pgxpool.Pool
}

func NewReviewRepo(db *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (r *ReviewRepo) Create(ctx context.Context, review *models.Review) error {
	logger.Debug(ctx, "creating review", logrus.Fields{"user_id": review.UserID, "place_id": review.PlaceID})
	query := `INSERT INTO review (user_id, place_id, title, rating, comment, visit_date)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query,
		review.UserID, review.PlaceID, review.Title, review.Rating, review.Comment, review.VisitDate,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		logger.Error(ctx, "failed to create review", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "review created", logrus.Fields{"review_id": review.ID})
	return nil
}

func (r *ReviewRepo) GetByID(ctx context.Context, id uint64) (*models.Review, error) {
	logger.Debug(ctx, "getting review by id", logrus.Fields{"review_id": id})
	query := `SELECT id, user_id, place_id, title, rating, comment, visit_date, created_at, updated_at
              FROM review WHERE id = $1`
	var rev models.Review
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rev.ID, &rev.UserID, &rev.PlaceID, &rev.Title, &rev.Rating, &rev.Comment, &rev.VisitDate,
		&rev.CreatedAt, &rev.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Debug(ctx, "review not found", logrus.Fields{"review_id": id})
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get review by id", logrus.Fields{"error": err})
		return nil, err
	}
	return &rev, nil
}

func (r *ReviewRepo) GetByPlaceIDWithAuthor(ctx context.Context, placeID uint64) ([]models.ReviewWithAuthor, error) {
	logger.Debug(ctx, "getting reviews with author", logrus.Fields{"place_id": placeID})
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
		logger.Error(ctx, "failed to get reviews with author", logrus.Fields{"error": err})
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
			logger.Error(ctx, "failed to scan review row", logrus.Fields{"error": err})
			return nil, err
		}
		rev.Author.ID = authorID
		rev.Author.Nickname = authorNickname
		rev.Author.Avatar = authorAvatar
		reviews = append(reviews, rev)
	}
	logger.Debug(ctx, "reviews retrieved", logrus.Fields{"count": len(reviews)})
	return reviews, nil
}

func (r *ReviewRepo) Delete(ctx context.Context, id uint64) error {
	logger.Debug(ctx, "deleting review", logrus.Fields{"review_id": id})
	_, err := r.db.Exec(ctx, `DELETE FROM review WHERE id = $1`, id)
	if err != nil {
		logger.Error(ctx, "failed to delete review", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "review deleted", logrus.Fields{"review_id": id})
	return nil
}
