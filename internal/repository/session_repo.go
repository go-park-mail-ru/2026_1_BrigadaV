package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"guidely-app/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepo(db *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *models.Session) error {
	hashedToken, err := utils.HashPassword(session.SessionToken)
	if err != nil {
		return err
	}
	query := `INSERT INTO session (user_id, session_token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, session.UserID, hashedToken, session.ExpiresAt).Scan(&session.ID, &session.CreatedAt)
}

func (r *SessionRepo) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	hashedToken, err := utils.HashPassword(token)
	if err != nil {
		return nil, err
	}
	query := `SELECT id, user_id, session_token_hash, expires_at, created_at FROM session WHERE session_token_hash = $1`
	var session models.Session
	var tokenHash string
	err = r.db.QueryRow(ctx, query, hashedToken).Scan(&session.ID, &session.UserID, &tokenHash, &session.ExpiresAt, &session.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &session, err
}

func (r *SessionRepo) DeleteByToken(ctx context.Context, token string) error {
	hashedToken, err := utils.HashPassword(token)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM session WHERE session_token_hash = $1`, hashedToken)
	return err
}
