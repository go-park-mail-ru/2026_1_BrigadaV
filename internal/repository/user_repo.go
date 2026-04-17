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

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	logger.Debug(ctx, "creating user", logrus.Fields{"login": user.Login})
	
	query := `INSERT INTO "user" (login, nickname, avatar_url, password_hash, country, city, about) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	
	err := r.db.QueryRow(ctx, query, user.Login, user.Nickname, user.AvatarURL, user.PasswordHash,
		user.Country, user.City, user.About).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logger.Error(ctx, "failed to create user", logrus.Fields{"error": err})
		return err
	}
	
	logger.Debug(ctx, "user created", logrus.Fields{"user_id": user.ID})
	return nil
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	logger.Debug(ctx, "getting user by login", logrus.Fields{"login": login})
	
	query := `SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at 
              FROM "user" WHERE login = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash,
		&user.Country, &user.City, &user.About, &user.HasReviews,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Debug(ctx, "user not found by login", logrus.Fields{"login": login})
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get user by login", logrus.Fields{"error": err})
		return nil, err
	}
	
	logger.Debug(ctx, "user found by login", logrus.Fields{"user_id": user.ID, "login": user.Login})
	return &user, nil
}

func (r *UserRepo) GetByNickname(ctx context.Context, nickname string) (*models.User, error) {
	logger.Debug(ctx, "getting user by nickname", logrus.Fields{"nickname": nickname})
	
	query := `SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at 
              FROM "user" WHERE nickname = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, nickname).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash,
		&user.Country, &user.City, &user.About, &user.HasReviews,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Debug(ctx, "user not found by nickname", logrus.Fields{"nickname": nickname})
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get user by nickname", logrus.Fields{"error": err})
		return nil, err
	}
	
	logger.Debug(ctx, "user found by nickname", logrus.Fields{"user_id": user.ID, "nickname": user.Nickname})
	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	logger.Debug(ctx, "getting user by id", logrus.Fields{"user_id": id})
	query := `SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at 
              FROM "user" WHERE id = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Login, &user.Nickname, &user.AvatarURL, &user.PasswordHash,
		&user.Country, &user.City, &user.About, &user.HasReviews,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Debug(ctx, "user not found by id", logrus.Fields{"user_id": id})
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get user by id", logrus.Fields{"error": err})
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	logger.Debug(ctx, "updating user", logrus.Fields{"user_id": user.ID})
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
	err := r.db.QueryRow(ctx, query,
		user.Login, user.Nickname, user.AvatarURL, user.Country, user.City, user.About, user.HasReviews, user.ID,
	).Scan(&user.UpdatedAt)
	if err != nil {
		logger.Error(ctx, "failed to update user", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "user updated", logrus.Fields{"user_id": user.ID})
	return nil
}
