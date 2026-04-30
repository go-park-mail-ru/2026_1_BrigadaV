package repository

import (
	"context"
	"errors"
	"fmt"
	"guidely-app/internal/logger"
	"guidely-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type PlaceRepo struct {
	db DB
}

func NewPlaceRepo(db DB) *PlaceRepo {
	return &PlaceRepo{db: db}
}

func (r *PlaceRepo) GetAll(ctx context.Context) ([]models.Place, error) {
	logger.Debug(ctx, "getting all places", nil)
	query := `
        SELECT p.id, p.name, p.description, p.photo_url, p.price, p.created_at, p.updated_at,
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description,
               pp.id as place_photo_id, ph.file_path, pp.is_main
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        LEFT JOIN place_photo pp ON p.id = pp.place_id
        LEFT JOIN photo ph ON pp.photo_id = ph.id
        ORDER BY p.id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.Error(ctx, "failed to get all places", logrus.Fields{"error": err})
		return nil, err
	}
	defer rows.Close()

	placesMap := make(map[uint64]*models.Place)

	for rows.Next() {
		var p models.Place
		var locID, catID *uint64
		var locName, countryName *string
		var locLat, locLng *float64
		var catName, catDesc *string
		var placePhotoID *uint64
		var photoFilePath *string
		var isMain *bool

		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.Price, &p.CreatedAt, &p.UpdatedAt,
			&locID, &locName, &countryName, &locLat, &locLng,
			&catID, &catName, &catDesc,
			&placePhotoID, &photoFilePath, &isMain,
		)
		if err != nil {
			logger.Error(ctx, "failed to scan place row", logrus.Fields{"error": err})
			return nil, err
		}

		if _, exists := placesMap[p.ID]; !exists {
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
			placesMap[p.ID] = &p
		}

		if placePhotoID != nil && photoFilePath != nil {
			photo := models.PlacePhoto{
				ID:      *placePhotoID,
				PlaceID: p.ID,
				PhotoID: *placePhotoID,
				Photo: models.Photo{
					ID:       *placePhotoID,
					FilePath: *photoFilePath,
				},
				IsMain: isMain != nil && *isMain,
			}
			placesMap[p.ID].Photos = append(placesMap[p.ID].Photos, photo)
		}
	}

	var places []models.Place
	for _, p := range placesMap {
		places = append(places, *p)
	}

	logger.Debug(ctx, "places retrieved", logrus.Fields{"count": len(places)})
	return places, nil
}

func (r *PlaceRepo) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	logger.Debug(ctx, "getting place by id", logrus.Fields{"place_id": id})
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
		logger.Debug(ctx, "place not found", logrus.Fields{"place_id": id})
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "failed to get place by id", logrus.Fields{"error": err})
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
	var place models.Place
	query := `SELECT id, name, description, photo_url, price, rating, review_count FROM place WHERE id = $1`
	err := r.db.QueryRow(ctx, query, placeID).Scan(&place.ID, &place.Name, &place.Description, &place.PhotoURL, &place.Price, &place.Rating, &place.ReviewCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get place: %w", err)
	}

	var isLiked bool
	if userID != 0 {
		_ = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorite WHERE user_id=$1 AND place_id=$2)`, userID, placeID).Scan(&isLiked)
	}
	return &models.PlaceWithRating{
		ID:          place.ID,
		Name:        place.Name,
		Description: place.Description,
		PhotoURL:    place.PhotoURL,
		Price:       place.Price,
		Rating:      place.Rating,
		ReviewCount: int64(place.ReviewCount),
		IsLiked:     isLiked,
	}, nil
}

func (r *PlaceRepo) IsPlaceInTrip(ctx context.Context, placeID, tripID uint64) (bool, error) {
	logger.Debug(ctx, "checking if place is in trip", logrus.Fields{"place_id": placeID, "trip_id": tripID})
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM trip_attractions 
			WHERE place_id = $1 AND trip_id = $2
		)`, placeID, tripID).Scan(&exists)
	if err != nil {
		logger.Error(ctx, "failed to check place in trip", logrus.Fields{"error": err})
		return false, err
	}
	return exists, nil
}
