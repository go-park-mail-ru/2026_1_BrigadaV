package handlers

import (
	"encoding/json"
	"guidely-app/internal/models"
	"guidely-app/internal/service"
	"net/http"
)

type PlaceHandler struct {
	placeService *service.PlaceService
}

func NewPlaceHandler(placeService *service.PlaceService) *PlaceHandler {
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

	response := make([]models.PlaceResponse, 0, len(places))
	for _, p := range places {
		liked := false
		pr := models.PlaceResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			IsLiked:     liked,
			Locality:    p.Locality,
			CreatedAt:   p.CreatedAt,
		}
		if p.Category.ID != 0 {
			pr.Category = &p.Category
		}
		if len(p.Photos) > 0 {
			pr.Photos = p.Photos
		}
		response = append(response, pr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
