package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO "user" (email, nickname, avatar_url, password_hash) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, user.Login, user.Nickname, user.AvatarURL, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, nickname, avatar_url, password_hash, created_at, updated_at 
              FROM "user" WHERE email = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	query := `SELECT id, email, nickname, avatar_url, password_hash, created_at, updated_at 
              FROM "user" WHERE id = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `UPDATE "user" SET nickname = $1, avatar_url = $2, updated_at = NOW() 
              WHERE id = $3 RETURNING updated_at`
	return r.db.QueryRow(ctx, query, user.Nickname, user.AvatarURL, user.ID).Scan(&user.UpdatedAt)
}
