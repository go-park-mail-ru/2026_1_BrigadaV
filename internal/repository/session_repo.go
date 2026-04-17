package repository

import (
	"context"
	"errors"
	"guidely-app/internal/logger"
	"guidely-app/internal/models"
	"guidely-app/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type SessionRepo struct {
	db DB
}

func NewSessionRepo(db DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *models.Session) error {
	logger.Debug(ctx, "creating session", logrus.Fields{"user_id": session.UserID})
	hashedToken := utils.HashToken(session.SessionToken)
	query := `INSERT INTO session (user_id, session_token_hash, expires_at) 
              VALUES ($1, $2, $3) RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query, session.UserID, hashedToken, session.ExpiresAt).
		Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		logger.Error(ctx, "failed to create session", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "session created", logrus.Fields{"session_id": session.ID})
	return nil
}

func (r *SessionRepo) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	logger.Debug(ctx, "getting session by token", nil)
	hashedToken := utils.HashToken(token)
	query := `SELECT id, user_id, session_token_hash, expires_at, created_at 
              FROM session WHERE session_token_hash = $1`
	var session models.Session
	var tokenHash string
	err := r.db.QueryRow(ctx, query, hashedToken).Scan(
		&session.ID, &session.UserID, &tokenHash, &session.ExpiresAt, &session.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Debug(ctx, "session not found by token", nil)
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get session by token", logrus.Fields{"error": err})
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) DeleteByToken(ctx context.Context, token string) error {
	logger.Debug(ctx, "deleting session by token", nil)
	hashedToken := utils.HashToken(token)
	_, err := r.db.Exec(ctx, `DELETE FROM session WHERE session_token_hash = $1`, hashedToken)
	if err != nil {
		logger.Error(ctx, "failed to delete session by token", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "session deleted", nil)
	return nil
}
