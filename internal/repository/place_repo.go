package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"guidely-app/internal/logger"
	"guidely-app/pkg/models"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type PlaceRepo struct {
	db DB
}

func NewPlaceRepo(db DB) *PlaceRepo {
	return &PlaceRepo{db: db}
}

const placeSelectCols = `
	p.id, p.name, p.description, p.photo_url, p.price, p.created_at, p.updated_at,
	p.latitude, p.longitude,
	COALESCE(p.rating, 0), COALESCE(p.review_count, 0),
	l.id, l.name, c.name AS country_name, l.latitude, l.longitude,
	cat.id, cat.name, cat.description,
	pp.id AS place_photo_id, ph.file_path, pp.is_main`

const placeJoins = `
	FROM place p
	LEFT JOIN locality l ON p.locality_id = l.id
	LEFT JOIN country c ON l.country_id = c.id
	LEFT JOIN category cat ON p.category_id = cat.id
	LEFT JOIN place_photo pp ON p.id = pp.place_id
	LEFT JOIN photo ph ON pp.photo_id = ph.id`

func buildFilterClauses(filter PlaceFilter, args []any) (string, []any) {
	var conditions []string

	if len(filter.CategoryIDs) > 0 {
		placeholders := make([]string, len(filter.CategoryIDs))
		for i, id := range filter.CategoryIDs {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		conditions = append(conditions, fmt.Sprintf("p.category_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.MinRating > 0 {
		args = append(args, filter.MinRating)
		conditions = append(conditions, fmt.Sprintf("COALESCE(p.rating, 0) >= $%d", len(args)))
	}

	if filter.MinReviews > 0 {
		args = append(args, filter.MinReviews)
		conditions = append(conditions, fmt.Sprintf("COALESCE(p.review_count, 0) >= $%d", len(args)))
	}

	clause := ""
	if len(conditions) > 0 {
		clause = " AND " + strings.Join(conditions, " AND ")
	}
	return clause, args
}

func (r *PlaceRepo) queryPlaces(ctx context.Context, query string, args ...any) ([]models.Place, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	placesMap := make(map[uint64]*models.Place)
	var order []uint64

	for rows.Next() {
		var p models.Place
		var locID, catID *uint64
		var locName, countryName *string
		var locLat, locLng *float64
		var catName, catDesc *string
		var placePhotoID *uint64
		var photoFilePath *string
		var isMain *bool

		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.Price, &p.CreatedAt, &p.UpdatedAt,
			&p.Latitude, &p.Longitude,
			&p.Rating, &p.ReviewCount,
			&locID, &locName, &countryName, &locLat, &locLng,
			&catID, &catName, &catDesc,
			&placePhotoID, &photoFilePath, &isMain,
		); err != nil {
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
			order = append(order, p.ID)
		}

		if placePhotoID != nil && photoFilePath != nil {
			photo := models.PlacePhoto{
				ID:      *placePhotoID,
				PlaceID: p.ID,
				PhotoID: *placePhotoID,
				Photo:   models.Photo{ID: *placePhotoID, FilePath: *photoFilePath},
				IsMain:  isMain != nil && *isMain,
			}
			placesMap[p.ID].Photos = append(placesMap[p.ID].Photos, photo)
		}
	}

	places := make([]models.Place, 0, len(order))
	for _, id := range order {
		places = append(places, *placesMap[id])
	}
	return places, nil
}

func (r *PlaceRepo) GetAll(ctx context.Context, filter PlaceFilter) ([]models.Place, error) {
	logger.Debug(ctx, "getting all places", nil)

	args := []any{}
	filterClause, args := buildFilterClauses(filter, args)

	query := fmt.Sprintf(
		"SELECT %s %s WHERE 1=1 %s ORDER BY p.id",
		placeSelectCols, placeJoins, filterClause,
	)

	places, err := r.queryPlaces(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "failed to get all places", logrus.Fields{"error": err})
		return nil, err
	}

	logger.Debug(ctx, "places retrieved", logrus.Fields{"count": len(places)})
	return places, nil
}

func (r *PlaceRepo) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	logger.Debug(ctx, "getting place by id", logrus.Fields{"place_id": id})
	query := `
        SELECT p.id, p.name, p.description, p.photo_url, p.price, p.created_at, p.updated_at,
               p.latitude, p.longitude,
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
		&p.Latitude, &p.Longitude,
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
	query := `SELECT id, name, description, photo_url, price, rating, review_count, latitude, longitude
	          FROM place WHERE id = $1`
	err := r.db.QueryRow(ctx, query, placeID).Scan(
		&place.ID, &place.Name, &place.Description, &place.PhotoURL, &place.Price,
		&place.Rating, &place.ReviewCount,
		&place.Latitude, &place.Longitude,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get place: %w", err)
	}

	var isLiked bool
	if userID != 0 {
		_ = r.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM favorite WHERE user_id=$1 AND place_id=$2)`,
			userID, placeID,
		).Scan(&isLiked)
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
		Latitude:    place.Latitude,
		Longitude:   place.Longitude,
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

func (r *PlaceRepo) GetByCategory(ctx context.Context, categoryID uint64) ([]models.Place, error) {
	logger.Debug(ctx, "getting places by category", logrus.Fields{"category_id": categoryID})
	query := fmt.Sprintf(
		"SELECT %s %s WHERE p.category_id = $1 ORDER BY p.id",
		placeSelectCols, placeJoins,
	)
	places, err := r.queryPlaces(ctx, query, categoryID)
	if err != nil {
		logger.Error(ctx, "failed to get places by category", logrus.Fields{"error": err})
		return nil, err
	}
	logger.Debug(ctx, "places by category retrieved", logrus.Fields{"count": len(places)})
	return places, nil
}

func (r *PlaceRepo) Search(ctx context.Context, query string, filter PlaceFilter) ([]models.Place, error) {
	logger.Debug(ctx, "searching places", logrus.Fields{"query": query})

	pattern := "%" + query + "%"
	args := []any{pattern}
	filterClause, args := buildFilterClauses(filter, args)

	q := fmt.Sprintf(
		"SELECT %s %s WHERE (p.name ILIKE $1 OR p.description ILIKE $1) %s ORDER BY p.id",
		placeSelectCols, placeJoins, filterClause,
	)

	places, err := r.queryPlaces(ctx, q, args...)
	if err != nil {
		logger.Error(ctx, "failed to search places", logrus.Fields{"error": err})
		return nil, err
	}

	logger.Debug(ctx, "search completed", logrus.Fields{"count": len(places)})
	return places, nil
}
