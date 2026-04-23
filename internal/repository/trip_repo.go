package repository

import (
	"context"
	"errors"
	"guidely-app/internal/logger"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type TripRepo struct {
	db DB
}

func NewTripRepo(db DB) *TripRepo {
	return &TripRepo{db: db}
}

func (r *TripRepo) Create(ctx context.Context, trip *models.Trip) error {
	logger.Debug(ctx, "creating trip", logrus.Fields{"title": trip.Title})
	query := `INSERT INTO trip (title, description, location, start_date, end_date, preview_url, created_by, is_public)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query,
		trip.Title, trip.Description, trip.Location, trip.StartDate, trip.EndDate, trip.PreviewURL,
		trip.CreatedBy, trip.IsPublic,
	).Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)
	if err != nil {
		logger.Error(ctx, "failed to create trip", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "trip created", logrus.Fields{"trip_id": trip.ID})
	return nil
}

func (r *TripRepo) GetByID(ctx context.Context, id uint64) (*models.Trip, error) {
	logger.Debug(ctx, "getting trip by id", logrus.Fields{"trip_id": id})
	query := `SELECT id, title, description, location, start_date, end_date, preview_url, created_by, is_public, created_at, updated_at
              FROM trip WHERE id = $1`
	var trip models.Trip
	err := r.db.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.Title, &trip.Description, &trip.Location, &trip.StartDate, &trip.EndDate, &trip.PreviewURL,
		&trip.CreatedBy, &trip.IsPublic, &trip.CreatedAt, &trip.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Debug(ctx, "trip not found", logrus.Fields{"trip_id": id})
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get trip by id", logrus.Fields{"error": err})
		return nil, err
	}
	return &trip, err
}

func (r *TripRepo) GetByUser(ctx context.Context, userID uint64) ([]models.Trip, error) {
	logger.Debug(ctx, "getting trips by user", logrus.Fields{"user_id": userID})
	query := `SELECT id, title, description, location, start_date, end_date, preview_url, created_by, is_public, created_at, updated_at
              FROM trip WHERE created_by = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		logger.Error(ctx, "failed to get trips by user", logrus.Fields{"error": err})
		return nil, err
	}
	defer rows.Close()
	var trips []models.Trip
	for rows.Next() {
		var t models.Trip
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Location, &t.StartDate, &t.EndDate, &t.PreviewURL,
			&t.CreatedBy, &t.IsPublic, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			logger.Error(ctx, "failed to scan trip row", logrus.Fields{"error": err})
			return nil, err
		}
		trips = append(trips, t)
	}
	logger.Debug(ctx, "trips retrieved", logrus.Fields{"count": len(trips)})
	return trips, nil
}

func (r *TripRepo) Update(ctx context.Context, trip *models.Trip) error {
	logger.Debug(ctx, "updating trip", logrus.Fields{"trip_id": trip.ID})
	query := `UPDATE trip SET 
        title = $1, description = $2, location = $3, start_date = $4, end_date = $5, preview_url = $6, 
        is_public = $7, updated_at = NOW() 
        WHERE id = $8 RETURNING updated_at`
	err := r.db.QueryRow(ctx, query,
		trip.Title, trip.Description, trip.Location, trip.StartDate, trip.EndDate, trip.PreviewURL,
		trip.IsPublic, trip.ID,
	).Scan(&trip.UpdatedAt)
	if err != nil {
		logger.Error(ctx, "failed to update trip", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "trip updated", logrus.Fields{"trip_id": trip.ID})
	return nil
}

func (r *TripRepo) Delete(ctx context.Context, id uint64) error {
	logger.Debug(ctx, "deleting trip", logrus.Fields{"trip_id": id})
	_, err := r.db.Exec(ctx, `DELETE FROM trip WHERE id = $1`, id)
	if err != nil {
		logger.Error(ctx, "failed to delete trip", logrus.Fields{"error": err})
		return err
	}
	logger.Debug(ctx, "trip deleted", logrus.Fields{"trip_id": id})
	return nil
}

func (r *TripRepo) AddAttraction(ctx context.Context, tripID, placeID uint64, order int16) error {
	logger.Debug(ctx, "adding attraction to trip", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	_, err := r.db.Exec(ctx, `INSERT INTO trip_attractions (trip_id, place_id, order_index) VALUES ($1, $2, $3)`, tripID, placeID, order)
	if err != nil {
		logger.Error(ctx, "failed to add attraction", logrus.Fields{"error": err})
		return err
	}
	return nil
}

func (r *TripRepo) GetAttractions(ctx context.Context, tripID uint64) ([]models.PlaceInTrip, error) {
	logger.Debug(ctx, "getting attractions for trip", logrus.Fields{"trip_id": tripID})
	query := `
        SELECT p.id, p.name, p.description,
               COALESCE(AVG(r.rating), 0) as rating,
               p.photo_url
        FROM trip_attractions ta
        JOIN place p ON ta.place_id = p.id
        LEFT JOIN review r ON r.place_id = p.id
        WHERE ta.trip_id = $1
        GROUP BY p.id, p.name, p.description, p.photo_url, ta.order_index
        ORDER BY ta.order_index
    `
	rows, err := r.db.Query(ctx, query, tripID)
	if err != nil {
		logger.Error(ctx, "failed to get attractions", logrus.Fields{"error": err})
		return nil, err
	}
	defer rows.Close()

	var places []models.PlaceInTrip
	for rows.Next() {
		var pl models.PlaceInTrip
		if err := rows.Scan(&pl.ID, &pl.Name, &pl.Description, &pl.Rating, &pl.PhotoURL); err != nil {
			logger.Error(ctx, "failed to scan attraction", logrus.Fields{"error": err})
			return nil, err
		}
		places = append(places, pl)
	}
	logger.Debug(ctx, "attractions retrieved", logrus.Fields{"count": len(places)})
	return places, nil
}

func (r *TripRepo) GetPlaceIDs(ctx context.Context, tripID uint64) ([]uint64, error) {
	logger.Debug(ctx, "getting place IDs for trip", logrus.Fields{"trip_id": tripID})
	query := `SELECT place_id FROM trip_attractions WHERE trip_id = $1 ORDER BY order_index`
	rows, err := r.db.Query(ctx, query, tripID)
	if err != nil {
		logger.Error(ctx, "failed to get place IDs", logrus.Fields{"error": err})
		return nil, err
	}
	defer rows.Close()

	var placeIDs []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			logger.Error(ctx, "failed to scan place ID", logrus.Fields{"error": err})
			return nil, err
		}
		placeIDs = append(placeIDs, id)
	}

	logger.Debug(ctx, "place IDs retrieved", logrus.Fields{"count": len(placeIDs)})
	return placeIDs, nil
}

func (r *TripRepo) RemoveAttraction(ctx context.Context, tripID, placeID uint64) error {
	logger.Debug(ctx, "removing attraction from trip", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	query := `DELETE FROM trip_attractions WHERE trip_id = $1 AND place_id = $2`
	_, err := r.db.Exec(ctx, query, tripID, placeID)
	if err != nil {
		logger.Error(ctx, "failed to remove attraction", logrus.Fields{"error": err})
		return err
	}
	return nil
}

func (r *TripRepo) CheckPlaceInTrip(ctx context.Context, tripID, placeID uint64) (bool, error) {
	logger.Debug(ctx, "checking if place is in trip", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	query := `SELECT EXISTS(SELECT 1 FROM trip_attractions WHERE trip_id = $1 AND place_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, tripID, placeID).Scan(&exists)
	if err != nil {
		logger.Error(ctx, "failed to check place in trip", logrus.Fields{"error": err})
		return false, err
	}
	return exists, nil
}
