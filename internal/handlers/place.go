package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"guidely-app/pkg/models"

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

// ratingThresholds maps frontend rating filter IDs to minimum rating values.
// IDs come from the RatingAccordion widget in the frontend Filters component.
var ratingThresholds = map[int]float64{
	1: 4.5,
	2: 4.0,
	3: 3.5,
	4: 3.0,
	5: 2.5,
}

func parseFilter(r *http.Request) service.PlaceFilter {
	filter := service.PlaceFilter{}

	if raw := r.URL.Query().Get("category_ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.ParseUint(part, 10, 64); err == nil {
				filter.CategoryIDs = append(filter.CategoryIDs, id)
			}
		}
	}

	if raw := r.URL.Query().Get("rating_ids"); raw != "" {
		var minRating float64
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.Atoi(part); err == nil {
				if threshold, ok := ratingThresholds[id]; ok && threshold < minRating || minRating == 0 {
					minRating = threshold
				}
			}
		}
		filter.MinRating = minRating
	}

	if raw := r.URL.Query().Get("min_reviews"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			filter.MinReviews = n
		}
	}

	return filter
}

func placeToDTO(p models.Place) dto.PlaceResponse {
	pr := dto.PlaceResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		IsLiked:     false,
		Locality: dto.LocalityDTO{
			ID:        p.Locality.ID,
			Name:      p.Locality.Name,
			Country:   p.Locality.Country,
			Latitude:  p.Locality.Latitude,
			Longitude: p.Locality.Longitude,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		PhotoURL:  p.PhotoURL,
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

	return pr
}

func placesToDTO(places []models.Place) []dto.PlaceResponse {
	response := make([]dto.PlaceResponse, 0, len(places))
	for _, p := range places {
		response = append(response, placeToDTO(p))
	}
	return response
}

func (h *PlaceHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := parseFilter(r)

	places, err := h.placeService.GetAll(r.Context(), filter)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch places", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch places"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(placesToDTO(places))
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
		if err.Error() == "place not found" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "place not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(place)
}

func (h *PlaceHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	placeID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid place id", http.StatusBadRequest)
		return
	}
	reviews, err := h.placeService.GetReviews(r.Context(), placeID)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch reviews", logrus.Fields{"place_id": placeID, "error": err})
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip_id"})
		return
	}

	trip, _, err := h.tripService.GetTripDetails(r.Context(), tripID)
	if err != nil || trip == nil || trip.CreatedBy != userID {
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

func (h *PlaceHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing query parameter 'q'"})
		return
	}

	filter := parseFilter(r)

	places, err := h.placeService.Search(r.Context(), query, filter)
	if err != nil {
		logger.Error(r.Context(), "Failed to search places", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to search places"})
		return
	}

	result := placesToDTO(places)
	if result == nil {
		result = []dto.PlaceResponse{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
