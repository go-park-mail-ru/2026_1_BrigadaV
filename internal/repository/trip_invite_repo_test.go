package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestTripInviteRepo_CreateInvite_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	invite := &models.TripInvite{
		TripID:    1,
		Token:     "test_token",
		Role:      "editor",
		IsOneTime: true,
		CreatedBy: 100,
	}

	rows := mockPool.NewRows([]string{"id", "created_at"}).AddRow(uint64(10), time.Now())
	mockPool.ExpectQuery(`INSERT INTO trip_invite`).
		WithArgs(invite.TripID, invite.Token, invite.Role, invite.IsOneTime, invite.ExpiresAt, invite.CreatedBy).
		WillReturnRows(rows)

	err = repo.CreateInvite(context.Background(), invite)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), invite.ID)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_CreateInvite_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	invite := &models.TripInvite{TripID: 1, Token: "token", Role: "viewer", IsOneTime: false, CreatedBy: 1}
	// Указываем аргументы, которые ожидаются в запросе
	mockPool.ExpectQuery(`INSERT INTO trip_invite`).
		WithArgs(invite.TripID, invite.Token, invite.Role, invite.IsOneTime, invite.ExpiresAt, invite.CreatedBy).
		WillReturnError(errors.New("db error"))

	err = repo.CreateInvite(context.Background(), invite)
	assert.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_GetInviteByToken_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	now := time.Now()
	rows := mockPool.NewRows([]string{"id", "trip_id", "token", "role", "is_one_time", "expires_at", "used_at", "created_at", "created_by"}).
		AddRow(uint64(1), uint64(2), "token123", "editor", true, nil, nil, now, uint64(10))

	mockPool.ExpectQuery(`SELECT id, trip_id, token, role, is_one_time, expires_at, used_at, created_at, created_by FROM trip_invite WHERE token = \$1`).
		WithArgs("token123").
		WillReturnRows(rows)

	invite, err := repo.GetInviteByToken(context.Background(), "token123")
	assert.NoError(t, err)
	assert.NotNil(t, invite)
	assert.Equal(t, uint64(1), invite.ID)
	assert.Equal(t, "editor", invite.Role)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_GetInviteByToken_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	mockPool.ExpectQuery(`SELECT .+ FROM trip_invite WHERE token = \$1`).
		WithArgs("nonexistent").
		WillReturnError(pgx.ErrNoRows)

	invite, err := repo.GetInviteByToken(context.Background(), "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, invite)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_GetInviteByToken_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	mockPool.ExpectQuery(`SELECT .+ FROM trip_invite WHERE token = \$1`).
		WithArgs("token").
		WillReturnError(errors.New("db error"))

	invite, err := repo.GetInviteByToken(context.Background(), "token")
	assert.Error(t, err)
	assert.Nil(t, invite)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_MarkUsed_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	mockPool.ExpectExec(`UPDATE trip_invite SET used_at = NOW\(\) WHERE id = \$1`).
		WithArgs(uint64(5)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.MarkUsed(context.Background(), 5)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_MarkUsed_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	mockPool.ExpectExec(`UPDATE trip_invite SET used_at = NOW\(\) WHERE id = \$1`).
		WithArgs(uint64(5)).
		WillReturnError(errors.New("update failed"))

	err = repo.MarkUsed(context.Background(), 5)
	assert.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_DeleteInvite_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM trip_invite WHERE id = \$1`).
		WithArgs(uint64(10)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteInvite(context.Background(), 10)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_DeleteInvite_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM trip_invite WHERE id = \$1`).
		WithArgs(uint64(10)).
		WillReturnError(errors.New("delete error"))

	err = repo.DeleteInvite(context.Background(), 10)
	assert.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_GetInvitesByTrip_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	now := time.Now()
	rows := mockPool.NewRows([]string{"id", "trip_id", "token", "role", "is_one_time", "expires_at", "used_at", "created_at", "created_by"}).
		AddRow(uint64(1), uint64(10), "token1", "viewer", false, nil, nil, now, uint64(100)).
		AddRow(uint64(2), uint64(10), "token2", "editor", true, nil, nil, now, uint64(100))

	mockPool.ExpectQuery(`SELECT id, trip_id, token, role, is_one_time, expires_at, used_at, created_at, created_by FROM trip_invite WHERE trip_id = \$1`).
		WithArgs(uint64(10)).
		WillReturnRows(rows)

	invites, err := repo.GetInvitesByTrip(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, invites, 2)
	assert.Equal(t, "token1", invites[0].Token)
	assert.Equal(t, "editor", invites[1].Role)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_GetInvitesByTrip_Empty(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	rows := mockPool.NewRows([]string{"id", "trip_id", "token", "role", "is_one_time", "expires_at", "used_at", "created_at", "created_by"})
	mockPool.ExpectQuery(`SELECT .+ FROM trip_invite WHERE trip_id = \$1`).
		WithArgs(uint64(99)).
		WillReturnRows(rows)

	invites, err := repo.GetInvitesByTrip(context.Background(), 99)
	assert.NoError(t, err)
	assert.Empty(t, invites)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripInviteRepo_GetInvitesByTrip_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripInviteRepo(mockPool)

	mockPool.ExpectQuery(`SELECT .+ FROM trip_invite WHERE trip_id = \$1`).
		WithArgs(uint64(10)).
		WillReturnError(errors.New("db error"))

	invites, err := repo.GetInvitesByTrip(context.Background(), 10)
	assert.Error(t, err)
	assert.Nil(t, invites)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
