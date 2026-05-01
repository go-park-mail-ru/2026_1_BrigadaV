package repository

import (
	"context"
	"guidely-app/pkg/models"
	"guidely-app/pkg/utils"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepo_Create(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()
	repo := NewSessionRepo(mockPool)

	session := &models.Session{
		UserID:    1,
		TokenHash: utils.HashToken("test_token"),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mockPool.ExpectQuery(`INSERT INTO session`).WithArgs(session.UserID, session.TokenHash, session.ExpiresAt).
		WillReturnRows(mockPool.NewRows([]string{"id", "created_at"}).AddRow(uint64(1), time.Now()))

	err := repo.Create(context.Background(), session)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestSessionRepo_GetByToken_Found(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewSessionRepo(mockPool)

	token := "test_token"
	hashedToken := utils.HashToken(token)

	rows := mockPool.NewRows([]string{"id", "user_id", "session_token_hash", "expires_at", "created_at"}).
		AddRow(uint64(1), uint64(1), hashedToken, time.Now().Add(7*24*time.Hour), time.Now())

	mockPool.ExpectQuery(`SELECT id, user_id, session_token_hash, expires_at, created_at FROM session WHERE session_token_hash = \$1`).
		WithArgs(hashedToken).
		WillReturnRows(rows)

	session, err := repo.GetByToken(context.Background(), token)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, uint64(1), session.UserID)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestSessionRepo_GetByToken_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewSessionRepo(mockPool)

	token := "nonexistent"
	hashedToken := utils.HashToken(token)

	mockPool.ExpectQuery(`SELECT id, user_id, session_token_hash, expires_at, created_at FROM session WHERE session_token_hash = \$1`).
		WithArgs(hashedToken).
		WillReturnRows(mockPool.NewRows([]string{"id", "user_id", "session_token_hash", "expires_at", "created_at"}))

	session, err := repo.GetByToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Nil(t, session)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestSessionRepo_DeleteByToken(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewSessionRepo(mockPool)

	token := "test_token"
	hashedToken := utils.HashToken(token)

	mockPool.ExpectExec(`DELETE FROM session WHERE session_token_hash = \$1`).
		WithArgs(hashedToken).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteByToken(context.Background(), token)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
