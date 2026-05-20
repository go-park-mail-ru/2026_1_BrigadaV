package repository

import (
	"context"
	"errors"
	"guidely-app/internal/logger"
	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type tripMemberRepo struct {
	db DB
}

func NewTripMemberRepo(db DB) TripMemberRepository {
	return &tripMemberRepo{db: db}
}

func (r *tripMemberRepo) AddMember(ctx context.Context, tripID, userID uint64, role string) error {
	logger.Debug(ctx, "adding trip member", logrus.Fields{"trip_id": tripID, "user_id": userID, "role": role})
	query := `INSERT INTO trip_member (trip_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (trip_id, user_id) DO UPDATE SET role = EXCLUDED.role`
	_, err := r.db.Exec(ctx, query, tripID, userID, role)
	if err != nil {
		logger.Error(ctx, "failed to add trip member", logrus.Fields{"error": err})
		return err
	}
	return nil
}

func (r *tripMemberRepo) RemoveMember(ctx context.Context, tripID, userID uint64) error {
	logger.Debug(ctx, "removing trip member", logrus.Fields{"trip_id": tripID, "user_id": userID})
	_, err := r.db.Exec(ctx, `DELETE FROM trip_member WHERE trip_id = $1 AND user_id = $2 AND role != 'owner'`, tripID, userID)
	if err != nil {
		logger.Error(ctx, "failed to remove trip member", logrus.Fields{"error": err})
		return err
	}
	return nil
}

func (r *tripMemberRepo) GetMemberRole(ctx context.Context, tripID, userID uint64) (string, error) {
	query := `SELECT role FROM trip_member WHERE trip_id = $1 AND user_id = $2`
	var role string
	err := r.db.QueryRow(ctx, query, tripID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil // нет роли -> не участник
	}
	if err != nil {
		logger.Error(ctx, "failed to get member role", logrus.Fields{"error": err})
		return "", err
	}
	return role, nil
}

func (r *tripMemberRepo) GetTripMembers(ctx context.Context, tripID uint64) ([]models.TripMember, error) {
	query := `SELECT trip_id, user_id, role, joined_at FROM trip_member WHERE trip_id = $1`
	rows, err := r.db.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []models.TripMember
	for rows.Next() {
		var m models.TripMember
		if err := rows.Scan(&m.TripID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *tripMemberRepo) HasEditPermission(ctx context.Context, tripID, userID uint64) (bool, error) {
	role, err := r.GetMemberRole(ctx, tripID, userID)
	if err != nil {
		return false, err
	}
	return role == "owner" || role == "editor", nil
}

func (r *tripMemberRepo) HasViewPermission(ctx context.Context, tripID, userID uint64) (bool, error) {
	role, err := r.GetMemberRole(ctx, tripID, userID)
	if err != nil {
		return false, err
	}
	return role != "", nil // любой участник имеет право просмотра
}
