package repository

import (
	"context"
	"errors"
	"guidely-app/internal/models"
	"time"

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
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description,
               ph.id, ph.place_id, ph.photo_id, ph.is_main, ph.created_at,
               pt.id, pt.file_path, pt.created_at
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        LEFT JOIN place_photo ph ON p.id = ph.place_id
        LEFT JOIN photo pt ON ph.photo_id = pt.id
        ORDER BY p.id`

    rows, err := r.db.Query(ctx, query)
    if err != nil {
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
        var phID, placePhotoPlaceID, photoID *uint64
        var isMain *bool
        var phCreatedAt *time.Time
        var ptID *uint64
        var filePath *string
        var ptCreatedAt *time.Time

        err := rows.Scan(
            &p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt,
            &locID, &locName, &countryName, &locLat, &locLng,
            &catID, &catName, &catDesc,
            &phID, &placePhotoPlaceID, &photoID, &isMain, &phCreatedAt,
            &ptID, &filePath, &ptCreatedAt,
        )
        if err != nil {
            return nil, err
        }

        // Check if we already have this place
        existingPlace, exists := placesMap[p.ID]
        if !exists {
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
            p.Photos = []models.PlacePhoto{}
            placesMap[p.ID] = &p
            existingPlace = placesMap[p.ID]
        }

        // Add photo if exists
        if phID != nil && ptID != nil {
            photo := models.Photo{
                ID:        *ptID,
                FilePath:  *filePath,
                CreatedAt: *ptCreatedAt,
            }
            
            placePhoto := models.PlacePhoto{
                ID:        *phID,
                PlaceID:   *placePhotoPlaceID,
                PhotoID:   *photoID,
                IsMain:    *isMain,
                CreatedAt: *phCreatedAt,
                Photo:     photo,
            }
            
            existingPlace.Photos = append(existingPlace.Photos, placePhoto)
        }
    }

    // Convert map to slice
    places := make([]models.Place, 0, len(placesMap))
    for _, p := range placesMap {
        places = append(places, *p)
    }
    
    return places, nil
}

func (r *PlaceRepo) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
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
		&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt, &p.UpdatedAt,
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
		return nil, err
	}

	var isLiked bool
	if userID != 0 {
		err = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorite WHERE user_id = $1 AND place_id = $2)`, userID, placeID).Scan(&isLiked)
		if err != nil {
			isLiked = false
		}
	}

	place, err := r.GetByID(ctx, placeID)
	if err != nil || place == nil {
		return nil, err
	}

	return &models.PlaceWithRating{
		ID:          place.ID,
		Name:        place.Name,
		Description: place.Description,
		Price:       place.Price,
		Rating:      rating,
		ReviewCount: reviewCount,
		IsLiked:     isLiked,
	}, nil
}

