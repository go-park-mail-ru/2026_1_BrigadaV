package handlers

import (
	"context"
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/models"
	"net/http"
	"strconv"
	"strings"
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

func (h *PlaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	id, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid place id"}`, http.StatusBadRequest)
		return
	}
	place, err := h.placeService.GetByID(r.Context(), id)
	if err != nil || place == nil {
		http.Error(w, `{"error":"place not found"}`, http.StatusNotFound)
		return
	}
	response := dto.PlaceResponse{
		ID:          place.ID,
		Name:        place.Name,
		Description: place.Description,
		Price:       place.Price,
		IsLiked:     false,
		Locality: dto.LocalityDTO{
			ID:        place.Locality.ID,
			Name:      place.Locality.Name,
			Country:   place.Locality.Country,
			Latitude:  place.Locality.Latitude,
			Longitude: place.Locality.Longitude,
		},
		CreatedAt: place.CreatedAt,
		UpdatedAt: place.UpdatedAt,
	}
	if place.Category.ID != 0 {
		response.Category = &dto.CategoryDTO{
			ID:          place.Category.ID,
			Name:        place.Category.Name,
			Description: place.Category.Description,
		}
	}
	if len(place.Photos) > 0 {
		response.Photos = make([]dto.PlacePhotoDTO, len(place.Photos))
		for i, ph := range place.Photos {
			response.Photos[i] = dto.PlacePhotoDTO{
				ID:       ph.ID,
				PlaceID:  ph.PlaceID,
				FilePath: ph.FilePath,
				IsMain:   ph.IsMain,
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
