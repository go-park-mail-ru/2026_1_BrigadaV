package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type PlaceHandler struct {
	placeService service.PlaceService
	tripService  service.TripService
}

func NewPlaceHandler(placeService service.PlaceService, tripService service.TripService) *PlaceHandler {
	return &PlaceHandler{
		placeService: placeService,
		tripService:  tripService,
	}
}

func (h *PlaceHandler) List(w http.ResponseWriter, r *http.Request) {
	places, err := h.placeService.GetAll(r.Context())
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch places", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch places"})
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
					FilePath: ph.Photo.FilePath,
					IsMain:   ph.IsMain,
				}
			}
		}
		response = append(response, pr)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PlaceHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid place id"})
		return
	}
	userIDVal := r.Context().Value("user_id")
	var userID uint64
	if userIDVal != nil {
		if id, ok := userIDVal.(uint64); ok {
			userID = id
		}
	}
	place, err := h.placeService.GetDetails(r.Context(), id, userID)
	if err != nil {
		log.Printf("place get details error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if place == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "place not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(place)
}

func (h *PlaceHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid place id in GetReviews", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid place id"})
		return
	}
	reviews, err := h.placeService.GetReviews(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch reviews", logrus.Fields{"place_id": id, "error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch reviews"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

func (h *PlaceHandler) CheckPlaceInTrip(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	vars := mux.Vars(r)
	placeID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid place id in CheckPlaceInTrip", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid place id"})
		return
	}

	tripIDStr := r.URL.Query().Get("trip_id")
	if tripIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing trip_id"})
		return
	}
	tripID, err := strconv.ParseUint(tripIDStr, 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip_id in CheckPlaceInTrip", logrus.Fields{"trip_id": tripIDStr, "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip_id"})
		return
	}

	trip, _, err := h.tripService.GetTripDetails(r.Context(), tripID)
	if err != nil || trip == nil || trip.CreatedBy != userID {
		logger.Error(r.Context(), "Access denied or trip not found in CheckPlaceInTrip", logrus.Fields{"trip_id": tripID, "user_id": userID})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "trip not found or access denied"})
		return
	}

	inTrip, err := h.placeService.IsPlaceInTrip(r.Context(), placeID, tripID)
	if err != nil {
		logger.Error(r.Context(), "Failed to check place in trip", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to check place in trip"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"in_trip": inTrip})
}
