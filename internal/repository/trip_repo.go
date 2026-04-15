package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepo struct {
	db *pgxpool.Pool
}

func NewTripRepo(db *pgxpool.Pool) *TripRepo {
	return &TripRepo{db: db}
}

func (r *TripRepo) Create(ctx context.Context, trip *models.Trip) error {
	query := `INSERT INTO trip (title, description, location, start_date, end_date, preview_url, created_by, is_public)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query,
		trip.Title, trip.Description, trip.Location, trip.StartDate, trip.EndDate, trip.PreviewURL,
		trip.CreatedBy, trip.IsPublic,
	).Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)
}

func (r *TripRepo) GetByID(ctx context.Context, id uint64) (*models.Trip, error) {
	query := `SELECT id, title, description, location, start_date, end_date, preview_url, created_by, is_public, created_at, updated_at
              FROM trip WHERE id = $1`
	var trip models.Trip
	err := r.db.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.Title, &trip.Description, &trip.Location, &trip.StartDate, &trip.EndDate, &trip.PreviewURL,
		&trip.CreatedBy, &trip.IsPublic, &trip.CreatedAt, &trip.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &trip, err
}

func (r *TripRepo) GetByUser(ctx context.Context, userID uint64) ([]models.Trip, error) {
	query := `SELECT id, title, description, location, start_date, end_date, preview_url, created_by, is_public, created_at, updated_at
              FROM trip WHERE created_by = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var trips []models.Trip
	for rows.Next() {
		var t models.Trip
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Location, &t.StartDate, &t.EndDate, &t.PreviewURL,
			&t.CreatedBy, &t.IsPublic, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		trips = append(trips, t)
	}
	return trips, nil
}

func (r *TripRepo) Update(ctx context.Context, trip *models.Trip) error {
	query := `UPDATE trip SET 
        title = $1, description = $2, location = $3, start_date = $4, end_date = $5, preview_url = $6, 
        is_public = $7, updated_at = NOW() 
        WHERE id = $8 RETURNING updated_at`
	return r.db.QueryRow(ctx, query,
		trip.Title, trip.Description, trip.Location, trip.StartDate, trip.EndDate, trip.PreviewURL,
		trip.IsPublic, trip.ID,
	).Scan(&trip.UpdatedAt)
}

func (r *TripRepo) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM trip WHERE id = $1`, id)
	return err
}

func (r *TripRepo) AddAttraction(ctx context.Context, tripID, placeID uint64, order int16) error {
	_, err := r.db.Exec(ctx, `INSERT INTO trip_attractions (trip_id, place_id, order_index) VALUES ($1, $2, $3)`, tripID, placeID, order)
	return err
}

func (r *TripRepo) GetAttractions(ctx context.Context, tripID uint64) ([]models.PlaceInTrip, error) {
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
        return nil, err
    }
    defer rows.Close()
    
    var places []models.PlaceInTrip
    for rows.Next() {
        var pl models.PlaceInTrip
        if err := rows.Scan(&pl.ID, &pl.Name, &pl.Description, &pl.Rating, &pl.PhotoURL); err != nil {
            return nil, err
        }
        places = append(places, pl)
    }
    return places, nil
}
