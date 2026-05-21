package repository

import (
	"context"
	"errors"
	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	db DB
}

func NewUserRepo(db DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO "user" (login, nickname, avatar_url, password_hash, country, city, about)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, user.Login, user.Nickname, user.AvatarURL,
		user.PasswordHash, user.Country, user.City, user.About).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return err
}

func (r *UserRepo) CreateOAuth(ctx context.Context, user *models.User) error {
	query := `INSERT INTO "user" (login, nickname, avatar_url, password_hash, yandex_id, country, city, about)
	          VALUES ($1, $2, $3, '', $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, user.Login, user.Nickname, user.AvatarURL,
		user.YandexID, user.Country, user.City, user.About).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, nickname, avatar_url, password_hash, yandex_id, country, city, about, has_reviews, role, created_at, updated_at
	          FROM "user" WHERE login = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash, &user.YandexID,
		&user.Country, &user.City, &user.About, &user.HasReviews, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) GetByNickname(ctx context.Context, nickname string) (*models.User, error) {
	query := `SELECT id, login, nickname, avatar_url, password_hash, yandex_id, country, city, about, has_reviews, role, created_at, updated_at
	          FROM "user" WHERE nickname = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, nickname).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash, &user.YandexID,
		&user.Country, &user.City, &user.About, &user.HasReviews, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	query := `SELECT id, login, nickname, avatar_url, password_hash, yandex_id, country, city, about, has_reviews, role, created_at, updated_at
	          FROM "user" WHERE id = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash, &user.YandexID,
		&user.Country, &user.City, &user.About, &user.HasReviews, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) GetByYandexID(ctx context.Context, yandexID string) (*models.User, error) {
	query := `SELECT id, login, nickname, avatar_url, password_hash, yandex_id, country, city, about, has_reviews, role, created_at, updated_at
	          FROM "user" WHERE yandex_id = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, yandexID).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash, &user.YandexID,
		&user.Country, &user.City, &user.About, &user.HasReviews, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `UPDATE "user" SET login=$1, nickname=$2, avatar_url=$3, country=$4, city=$5, about=$6, has_reviews=$7, role=$8, updated_at=NOW()
	          WHERE id=$9 RETURNING updated_at`
	return r.db.QueryRow(ctx, query, user.Login, user.Nickname, user.AvatarURL,
		user.Country, user.City, user.About, user.HasReviews, user.Role, user.ID).Scan(&user.UpdatedAt)
}
