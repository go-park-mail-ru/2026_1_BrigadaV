package repository

import (
	"context"
	"errors"
	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
)

type SessionRepo struct {
	db DB
}

func NewSessionRepo(db DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *models.Session) error {
	query := `INSERT INTO session (user_id, session_token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, session.UserID, session.TokenHash, session.ExpiresAt).Scan(&session.ID, &session.CreatedAt)
}

func (r *SessionRepo) GetByToken(ctx context.Context, tokenHash string) (*models.Session, error) {
	query := `SELECT id, user_id, session_token_hash, expires_at, created_at FROM session WHERE session_token_hash = $1`
	var session models.Session
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt, &session.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &session, err
}

func (r *SessionRepo) DeleteByToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM session WHERE session_token_hash = $1`, tokenHash)
	return err
}
