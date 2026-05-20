package repository

import (
	"context"
	"errors"
	"guidely-app/internal/logger"
	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type tripInviteRepo struct {
	db DB
}

func NewTripInviteRepo(db DB) TripInviteRepository {
	return &tripInviteRepo{db: db}
}

func (r *tripInviteRepo) CreateInvite(ctx context.Context, invite *models.TripInvite) error {
	query := `INSERT INTO trip_invite (trip_id, token, role, is_one_time, expires_at, created_by)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query, invite.TripID, invite.Token, invite.Role, invite.IsOneTime, invite.ExpiresAt, invite.CreatedBy).
		Scan(&invite.ID, &invite.CreatedAt)
	if err != nil {
		logger.Error(ctx, "failed to create invite", logrus.Fields{"error": err})
		return err
	}
	return nil
}

func (r *tripInviteRepo) GetInviteByToken(ctx context.Context, token string) (*models.TripInvite, error) {
	query := `SELECT id, trip_id, token, role, is_one_time, expires_at, used_at, created_at, created_by
              FROM trip_invite WHERE token = $1`
	var inv models.TripInvite
	err := r.db.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.TripID, &inv.Token, &inv.Role, &inv.IsOneTime, &inv.ExpiresAt, &inv.UsedAt, &inv.CreatedAt, &inv.CreatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get invite by token", logrus.Fields{"error": err})
		return nil, err
	}
	return &inv, nil
}

func (r *tripInviteRepo) MarkUsed(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `UPDATE trip_invite SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *tripInviteRepo) DeleteInvite(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM trip_invite WHERE id = $1`, id)
	return err
}

func (r *tripInviteRepo) GetInvitesByTrip(ctx context.Context, tripID uint64) ([]models.TripInvite, error) {
	rows, err := r.db.Query(ctx, `SELECT id, trip_id, token, role, is_one_time, expires_at, used_at, created_at, created_by
                                 FROM trip_invite WHERE trip_id = $1`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invites []models.TripInvite
	for rows.Next() {
		var inv models.TripInvite
		if err := rows.Scan(&inv.ID, &inv.TripID, &inv.Token, &inv.Role, &inv.IsOneTime, &inv.ExpiresAt, &inv.UsedAt, &inv.CreatedAt, &inv.CreatedBy); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, nil
}
