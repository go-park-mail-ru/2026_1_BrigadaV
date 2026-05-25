package handlers

import (
	"encoding/json"
	"html/template"
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
		Rating:      p.Rating,
		ReviewCount: p.ReviewCount,
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

// FilterByReviewsAndRating godoc
// @Summary      Filter places by reviews and rating
// @Description  Returns places filtered strictly by minimum rating and/or minimum review count, sorted by rating desc, review count desc
// @Tags         places
// @Produce      json
// @Param        min_rating    query    number  false  "Minimum average rating (e.g. 4.0)"
// @Param        min_reviews   query    int     false  "Minimum number of reviews (e.g. 10)"
// @Param        rating_ids    query    string  false  "Comma-separated rating tier IDs (1=4.5+, 2=4.0+, 3=3.5+, 4=3.0+, 5=2.5+)"
// @Success      200  {array}   dto.PlaceResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /places/filter [get]
func (h *PlaceHandler) FilterByReviewsAndRating(w http.ResponseWriter, r *http.Request) {
	filter := service.PlaceFilter{}

	if raw := r.URL.Query().Get("min_rating"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil && v >= 0 {
			filter.MinRating = v
		} else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid min_rating value"})
			return
		}
	}

	if raw := r.URL.Query().Get("min_reviews"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			filter.MinReviews = v
		} else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid min_reviews value"})
			return
		}
	}

	// Also support rating_ids shorthand (same tiers as in parseFilter)
	if raw := r.URL.Query().Get("rating_ids"); raw != "" && filter.MinRating == 0 {
		var minRating float64
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.Atoi(part); err == nil {
				if threshold, ok := ratingThresholds[id]; ok && (minRating == 0 || threshold < minRating) {
					minRating = threshold
				}
			}
		}
		filter.MinRating = minRating
	}

	if filter.MinRating == 0 && filter.MinReviews == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "at least one of min_rating, min_reviews, or rating_ids must be specified"})
		return
	}

	places, err := h.placeService.FilterByReviewsAndRating(r.Context(), filter)
	if err != nil {
		logger.Error(r.Context(), "Failed to filter places", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to filter places"})
		return
	}

	result := placesToDTO(places)
	if result == nil {
		result = []dto.PlaceResponse{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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
	filter := parseFilter(r)

	var places []models.Place
	var err error

	if query == "" {
		places, err = h.placeService.GetAll(r.Context(), filter)
	} else {
		places, err = h.placeService.Search(r.Context(), query, filter)
	}

	if err != nil {
		logger.Error(r.Context(), "Failed to search/filter places", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to search/filter places"})
		return
	}

	result := placesToDTO(places)
	if result == nil {
		result = []dto.PlaceResponse{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *PlaceHandler) GetBotPreview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid place id"})
		return
	}

	place, err := h.placeService.GetDetails(r.Context(), id, 0)
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

	photoURL := place.PhotoURL
	if photoURL != "" && !strings.HasPrefix(photoURL, "http") {
		photoURL = "https://guidely.ru" + photoURL
	}

	const tmplStr = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta property="og:type" content="website">
    <meta property="og:url" content="https://guidely.ru/attraction/{{.ID}}">
    <meta property="og:title" content="{{.Name}}">
    <meta property="og:description" content="{{.Description}}">
    <meta property="og:image" content="{{.PhotoURL}}">
    <meta name="twitter:card" content="summary_large_image">
</head>
<body></body>
</html>`

	t, err := template.New("preview").Parse(tmplStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := struct {
		ID          uint64
		Name        string
		Description string
		PhotoURL    string
	}{
		ID:          place.ID,
		Name:        place.Name,
		Description: place.Description,
		PhotoURL:    photoURL,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}
