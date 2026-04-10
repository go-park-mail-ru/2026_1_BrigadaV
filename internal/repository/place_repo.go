package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlaceRepo struct {
	db *pgxpool.Pool
}

func NewPlaceRepo(db *pgxpool.Pool) *PlaceRepo {
	return &PlaceRepo{db: db}
}

func (r *PlaceRepo) GetAll(ctx context.Context) ([]models.Place, error) {
	query := `
        SELECT p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
               l.id, l.name, l.country, l.latitude, l.longitude,
               c.id, c.name, c.description
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN category c ON p.category_id = c.id
        ORDER BY p.id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var places []models.Place
	for rows.Next() {
		var p models.Place
		var loc models.Locality
		var cat models.Category
		var locID, catID *uint64
		var locName, locCountry *string
		var locLat, locLng *float64
		var catName, catDesc *string
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt,
			&locID, &locName, &locCountry, &locLat, &locLng,
			&catID, &catName, &catDesc)
		if err != nil {
			return nil, err
		}
		if locID != nil {
			loc.ID = *locID
			loc.Name = *locName
			loc.Country = *locCountry
			loc.Latitude = locLat
			loc.Longitude = locLng
			p.Locality = loc
		}
		if catID != nil {
			cat.ID = *catID
			cat.Name = *catName
			cat.Description = *catDesc
			p.Category = cat
		}
		places = append(places, p)
	}
	return places, nil
}

func (r *PlaceRepo) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	query := `
        SELECT p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
               l.id, l.name, l.country, l.latitude, l.longitude,
               c.id, c.name, c.description
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN category c ON p.category_id = c.id
        WHERE p.id = $1`
	var p models.Place
	var loc models.Locality
	var cat models.Category
	var locID, catID *uint64
	var locName, locCountry *string
	var locLat, locLng *float64
	var catName, catDesc *string
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt,
		&locID, &locName, &locCountry, &locLat, &locLng,
		&catID, &catName, &catDesc)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if locID != nil {
		loc.ID = *locID
		loc.Name = *locName
		loc.Country = *locCountry
		loc.Latitude = locLat
		loc.Longitude = locLng
		p.Locality = loc
	}
	if catID != nil {
		cat.ID = *catID
		cat.Name = *catName
		cat.Description = *catDesc
		p.Category = cat
	}
	return &p, nil
}
