package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO "user" (login, nickname, avatar_url, password_hash, country, city, about) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, user.Login, user.Nickname, user.AvatarURL, user.PasswordHash,
		user.Country, user.City, user.About).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) GetByEmail(ctx context.Context, login string) (*models.User, error) {
	log.Println("GetByEmail called with email:", login)
	query := `SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at 
              FROM "user" WHERE login = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash,
		&user.Country, &user.City, &user.About, &user.HasReviews,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
        log.Println("Database error:", err)
        return nil, err
	}
	log.Println("User found:", user.Login) 
	return &user, err
}

func (r *UserRepo) GetByNickname(ctx context.Context, nickname string) (*models.User, error) {
	query := `SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at 
              FROM "user" WHERE nickname = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, nickname).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash,
		&user.Country, &user.City, &user.About, &user.HasReviews,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	query := `SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at 
              FROM "user" WHERE id = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash,
		&user.Country, &user.City, &user.About, &user.HasReviews,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `UPDATE "user" SET 
        login = $1,
        nickname = $2, 
        avatar_url = $3, 
        country = $4, 
        city = $5, 
        about = $6, 
        has_reviews = $7,
        updated_at = NOW() 
    WHERE id = $8 RETURNING updated_at`
	return r.db.QueryRow(ctx, query,
		user.Login, user.Nickname, user.AvatarURL, user.Country, user.City, user.About, user.HasReviews, user.ID,
	).Scan(&user.UpdatedAt)
}
