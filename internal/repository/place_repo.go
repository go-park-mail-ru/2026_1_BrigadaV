package repository

import (
	"context"
	"errors"
	"fmt"
	"guidely-app/internal/models"
	"log"

	"github.com/jackc/pgx/v5"
)

type PlaceRepo struct {
	db DB
}

func NewPlaceRepo(db DB) *PlaceRepo {
	return &PlaceRepo{db: db}
}

func (r *PlaceRepo) GetAll(ctx context.Context) ([]models.Place, error) {
	query := `
        SELECT p.id, p.name, p.description, p.photo_url, p.price, p.created_at, p.updated_at,
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        ORDER BY p.id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var places []models.Place
	for rows.Next() {
		var p models.Place
		var locID, catID *uint64
		var locName, countryName *string
		var locLat, locLng *float64
		var catName, catDesc *string
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.Price, &p.CreatedAt, &p.UpdatedAt,
			&locID, &locName, &countryName, &locLat, &locLng,
			&catID, &catName, &catDesc,
		)
		if err != nil {
			return nil, err
		}
		if locID != nil {
			p.Locality = models.Locality{
				ID:        *locID,
				Name:      *locName,
				Country:   *countryName,
				Latitude:  locLat,
				Longitude: locLng,
			}
		}
		if catID != nil {
			p.Category = models.Category{
				ID:          *catID,
				Name:        *catName,
				Description: *catDesc,
			}
		}
		places = append(places, p)
	}
	return places, nil
}

func (r *PlaceRepo) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	query := `
        SELECT p.id, p.name, p.description, p.photo_url, p.price, p.created_at, p.updated_at,
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        WHERE p.id = $1`
	var p models.Place
	var locID, catID *uint64
	var locName, countryName *string
	var locLat, locLng *float64
	var catName, catDesc *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.Price, &p.CreatedAt, &p.UpdatedAt,
		&locID, &locName, &countryName, &locLat, &locLng,
		&catID, &catName, &catDesc,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if locID != nil {
		p.Locality = models.Locality{
			ID:        *locID,
			Name:      *locName,
			Country:   *countryName,
			Latitude:  locLat,
			Longitude: locLng,
		}
	}
	if catID != nil {
		p.Category = models.Category{
			ID:          *catID,
			Name:        *catName,
			Description: *catDesc,
		}
	}
	return &p, nil
}

func (r *PlaceRepo) GetWithRatingAndLike(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error) {
	var rating float64
	var reviewCount int64
	err := r.db.QueryRow(ctx, `
        SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM review WHERE place_id = $1
    `, placeID).Scan(&rating, &reviewCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}
	var isLiked bool
	if userID != 0 {
		err = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorite WHERE user_id = $1 AND place_id = $2)`, userID, placeID).Scan(&isLiked)
		if err != nil {
			log.Printf("[WARN] GetWithRatingAndLike: ошибка проверки лайка - %v", err)
			isLiked = false
		}
	}
	place, err := r.GetByID(ctx, placeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get place: %w", err)
	}
	if place == nil {
		return nil, fmt.Errorf("place with id %d not found", placeID)
	}
	return &models.PlaceWithRating{
		ID:          place.ID,
		Name:        place.Name,
		Description: place.Description,
		PhotoURL:    place.PhotoURL,
		Price:       place.Price,
		Rating:      rating,
		ReviewCount: reviewCount,
		IsLiked:     isLiked,
	}, nil
}
