package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
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
    log.Println("[DEBUG] GetAll: начал выполнение")

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
        log.Printf("[ERROR] GetAll: ошибка запроса - %v", err)
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
            log.Printf("[ERROR] GetAll: ошибка сканирования - %v", err)
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
        ID:       *placePhotoID,
        PlaceID:  p.ID,
        PhotoID:  *placePhotoID,
        Photo: models.Photo{
            ID:       *placePhotoID,
            FilePath: *photoFilePath,
        },
        IsMain:   isMain != nil && *isMain,
    }
    placesMap[p.ID].Photos = append(placesMap[p.ID].Photos, photo)
}
    }

    var places []models.Place
    for _, p := range placesMap {
        places = append(places, *p)
    }

    log.Printf("[DEBUG] GetAll: успешно загружено %d мест", len(places))
    return places, nil
}

func (r *PlaceRepo) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	log.Printf("[DEBUG] GetByID: поиск места с ID=%d", id)

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
		log.Printf("[DEBUG] GetByID: место с ID=%d не найдено", id)
		return nil, nil
	}
	if err != nil {
		log.Printf("[ERROR] GetByID: ошибка выполнения запроса для ID=%d - %v", id, err)
		return nil, err
	}

	log.Printf("[DEBUG] GetByID: найдено место ID=%d, название=%s", p.ID, p.Name)

	if locID != nil {
		p.Locality = models.Locality{
			ID:        *locID,
			Name:      *locName,
			Country:   *countryName,
			Latitude:  locLat,
			Longitude: locLng,
		}
		log.Printf("[DEBUG] GetByID: загружена локация ID=%d", *locID)
	}
	if catID != nil {
		p.Category = models.Category{
			ID:          *catID,
			Name:        *catName,
			Description: *catDesc,
		}
		log.Printf("[DEBUG] GetByID: загружена категория ID=%d", *catID)
	}

	return &p, nil
}

func (r *PlaceRepo) GetWithRatingAndLike(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error) {
	log.Printf("[DEBUG] GetWithRatingAndLike: начало - placeID=%d, userID=%d", placeID, userID)

	var rating float64
	var reviewCount int64

	log.Println("[DEBUG] GetWithRatingAndLike: запрос рейтинга...")
	err := r.db.QueryRow(ctx, `
        SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM review WHERE place_id = $1
    `, placeID).Scan(&rating, &reviewCount)

	if err != nil {
		log.Printf("[ERROR] GetWithRatingAndLike: ошибка получения рейтинга - %v", err)
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}
	log.Printf("[DEBUG] GetWithRatingAndLike: рейтинг=%.2f, кол-во отзывов=%d", rating, reviewCount)

	var isLiked bool
	if userID != 0 {
		log.Printf("[DEBUG] GetWithRatingAndLike: проверка лайка для userID=%d", userID)
		err = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorite WHERE user_id = $1 AND place_id = $2)`, userID, placeID).Scan(&isLiked)
		if err != nil {
			log.Printf("[WARN] GetWithRatingAndLike: ошибка проверки лайка - %v", err)
			isLiked = false
		}
		log.Printf("[DEBUG] GetWithRatingAndLike: isLiked=%v", isLiked)
	}

	log.Printf("[DEBUG] GetWithRatingAndLike: получение места с ID=%d", placeID)
	place, err := r.GetByID(ctx, placeID)
	if err != nil {
		log.Printf("[ERROR] GetWithRatingAndLike: ошибка получения места - %v", err)
		return nil, fmt.Errorf("failed to get place: %w", err)
	}

	if place == nil {
		log.Printf("[WARN] GetWithRatingAndLike: место с ID=%d не найдено", placeID)
		return nil, fmt.Errorf("place with id %d not found", placeID)
	}

	log.Printf("[DEBUG] GetWithRatingAndLike: место найдено, название=%s", place.Name)

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
