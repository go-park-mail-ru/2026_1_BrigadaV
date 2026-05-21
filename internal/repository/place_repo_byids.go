package repository

import (
	"context"
	"fmt"
	"guidely-app/internal/logger"
	"guidely-app/pkg/models"
	"strings"

	"github.com/sirupsen/logrus"
)

func (r *PlaceRepo) GetByIDs(ctx context.Context, ids []uint64) ([]models.Place, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	q := fmt.Sprintf(`
        SELECT p.id, p.name, p.description, p.photo_url, p.price, p.created_at, p.updated_at,
               p.latitude, p.longitude,
               l.id, l.name, c.name as country_name, l.latitude, l.longitude,
               cat.id, cat.name, cat.description,
               pp.id as place_photo_id, ph.file_path, pp.is_main
        FROM place p
        LEFT JOIN locality l ON p.locality_id = l.id
        LEFT JOIN country c ON l.country_id = c.id
        LEFT JOIN category cat ON p.category_id = cat.id
        LEFT JOIN place_photo pp ON p.id = pp.place_id
        LEFT JOIN photo ph ON pp.photo_id = ph.id
        WHERE p.id IN (%s)
        ORDER BY p.id`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		logger.Error(ctx, "failed to get places by ids", logrus.Fields{"error": err})
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
			&p.Latitude, &p.Longitude,
			&locID, &locName, &countryName, &locLat, &locLng,
			&catID, &catName, &catDesc,
			&placePhotoID, &photoFilePath, &isMain,
		)
		if err != nil {
			logger.Error(ctx, "failed to scan place row in GetByIDs", logrus.Fields{"error": err})
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

	result := make([]models.Place, 0, len(ids))
	for _, id := range ids {
		if p, ok := placesMap[id]; ok {
			result = append(result, *p)
		}
	}
	return result, nil
}
