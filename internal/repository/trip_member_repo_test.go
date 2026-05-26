package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestTripMemberRepo_AddMember_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectExec(`INSERT INTO trip_member \(trip_id, user_id, role\) VALUES \(\$1, \$2, \$3\) ON CONFLICT .+ DO UPDATE`).
		WithArgs(uint64(1), uint64(2), "editor").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.AddMember(context.Background(), 1, 2, "editor")
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_AddMember_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectExec(`INSERT INTO trip_member`).
		WithArgs(uint64(1), uint64(2), "viewer").
		WillReturnError(errors.New("db error"))

	err = repo.AddMember(context.Background(), 1, 2, "viewer")
	assert.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_RemoveMember_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM trip_member WHERE trip_id = \$1 AND user_id = \$2 AND role != 'owner'`).
		WithArgs(uint64(10), uint64(20)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.RemoveMember(context.Background(), 10, 20)
	assert.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_RemoveMember_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM trip_member`).
		WithArgs(uint64(10), uint64(20)).
		WillReturnError(errors.New("delete error"))

	err = repo.RemoveMember(context.Background(), 10, 20)
	assert.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_GetMemberRole_Found(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	rows := mockPool.NewRows([]string{"role"}).AddRow("owner")
	mockPool.ExpectQuery(`SELECT role FROM trip_member WHERE trip_id = \$1 AND user_id = \$2`).
		WithArgs(uint64(5), uint64(10)).
		WillReturnRows(rows)

	role, err := repo.GetMemberRole(context.Background(), 5, 10)
	assert.NoError(t, err)
	assert.Equal(t, "owner", role)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_GetMemberRole_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectQuery(`SELECT role FROM trip_member WHERE trip_id = \$1 AND user_id = \$2`).
		WithArgs(uint64(5), uint64(10)).
		WillReturnError(pgx.ErrNoRows) // используем pgx.ErrNoRows

	role, err := repo.GetMemberRole(context.Background(), 5, 10)
	assert.NoError(t, err)
	assert.Empty(t, role)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_GetMemberRole_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectQuery(`SELECT role FROM trip_member`).
		WithArgs(uint64(5), uint64(10)).
		WillReturnError(errors.New("query error"))

	role, err := repo.GetMemberRole(context.Background(), 5, 10)
	assert.Error(t, err)
	assert.Empty(t, role)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_GetTripMembers_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	now := time.Now()
	rows := mockPool.NewRows([]string{"trip_id", "user_id", "role", "joined_at"}).
		AddRow(uint64(1), uint64(10), "owner", now).
		AddRow(uint64(1), uint64(20), "editor", now)

	mockPool.ExpectQuery(`SELECT trip_id, user_id, role, joined_at FROM trip_member WHERE trip_id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	members, err := repo.GetTripMembers(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, uint64(10), members[0].UserID)
	assert.Equal(t, "editor", members[1].Role)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_GetTripMembers_Empty(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	rows := mockPool.NewRows([]string{"trip_id", "user_id", "role", "joined_at"})
	mockPool.ExpectQuery(`SELECT trip_id, user_id, role, joined_at FROM trip_member WHERE trip_id = \$1`).
		WithArgs(uint64(999)).
		WillReturnRows(rows)

	members, err := repo.GetTripMembers(context.Background(), 999)
	assert.NoError(t, err)
	assert.Empty(t, members)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_GetTripMembers_DBError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectQuery(`SELECT .+ FROM trip_member WHERE trip_id = \$1`).
		WithArgs(uint64(1)).
		WillReturnError(errors.New("db error"))

	members, err := repo.GetTripMembers(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, members)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_HasEditPermission_True(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	rows := mockPool.NewRows([]string{"role"}).AddRow("editor")
	mockPool.ExpectQuery(`SELECT role FROM trip_member WHERE trip_id = \$1 AND user_id = \$2`).
		WithArgs(uint64(1), uint64(2)).
		WillReturnRows(rows)

	ok, err := repo.HasEditPermission(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_HasEditPermission_False(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	rows := mockPool.NewRows([]string{"role"}).AddRow("viewer")
	mockPool.ExpectQuery(`SELECT role FROM trip_member WHERE trip_id = \$1 AND user_id = \$2`).
		WithArgs(uint64(1), uint64(2)).
		WillReturnRows(rows)

	ok, err := repo.HasEditPermission(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestTripMemberRepo_HasViewPermission_True(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	rows := mockPool.NewRows([]string{"role"}).AddRow("viewer")
	mockPool.ExpectQuery(`SELECT role FROM trip_member WHERE trip_id = \$1 AND user_id = \$2`).
		WithArgs(uint64(1), uint64(2)).
		WillReturnRows(rows)

	ok, err := repo.HasViewPermission(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// TestTripMemberRepo_HasViewPermission_NoRole
func TestTripMemberRepo_HasViewPermission_NoRole(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewTripMemberRepo(mockPool)

	mockPool.ExpectQuery(`SELECT role FROM trip_member WHERE trip_id = \$1 AND user_id = \$2`).
		WithArgs(uint64(1), uint64(2)).
		WillReturnError(pgx.ErrNoRows)

	ok, err := repo.HasViewPermission(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.False(t, ok)
}
