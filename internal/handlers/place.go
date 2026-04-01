package handlers

import (
	"context"
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/models"
	"net/http"
)

type PlaceService interface {
	GetAll(ctx context.Context) ([]models.Place, error)
	GetByID(ctx context.Context, id uint64) (*models.Place, error)
}

type PlaceHandler struct {
	placeService PlaceService
}

func NewPlaceHandler(placeService PlaceService) *PlaceHandler {
	return &PlaceHandler{placeService: placeService}
}

func (h *PlaceHandler) List(w http.ResponseWriter, r *http.Request) {
	places, err := h.placeService.GetAll(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to fetch places"}`, http.StatusInternalServerError)
		return
	}

	userIDVal := r.Context().Value("user_id")
	var userID uint64
	if userIDVal != nil {
		if id, ok := userIDVal.(uint64); ok {
			userID = id
		}
	}
	_ = userID

	response := make([]dto.PlaceResponse, 0, len(places))
	for _, p := range places {
		liked := false
		pr := dto.PlaceResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			IsLiked:     liked,
			Locality: dto.LocalityDTO{
				ID:        p.Locality.ID,
				Name:      p.Locality.Name,
				Country:   p.Locality.Country,
				Latitude:  p.Locality.Latitude,
				Longitude: p.Locality.Longitude,
			},
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
		if p.Category.ID != 0 {
			pr.Category = &dto.CategoryDTO{
				ID:          p.Category.ID,
				Name:        p.Category.Name,
				Description: p.Category.Description,
			}
		}
		if len(p.Photos) > 0 {
			pr.Photos = make([]dto.PlacePhotoDTO, len(p.Photos))
			for i, ph := range p.Photos {
				pr.Photos[i] = dto.PlacePhotoDTO{
					ID:       ph.ID,
					PlaceID:  ph.PlaceID,
					FilePath: ph.FilePath,
					IsMain:   ph.IsMain,
				}
			}
		}
		response = append(response, pr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
