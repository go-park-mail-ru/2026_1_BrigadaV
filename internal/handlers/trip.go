package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"guidely-app/internal/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type TripHandler struct {
	tripService service.TripService
}

func NewTripHandler(tripService service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	trips, err := h.tripService.GetUserTrips(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch trips"})
		return
	}
	response := make([]dto.TripResponse, len(trips))
	for i, t := range trips {
		response[i] = dto.TripResponse{
			ID:          t.ID,
			Title:       t.Title,
			Location:    t.Location,
			StartDate:   t.StartDate,
			EndDate:     t.EndDate,
			Description: t.Description,
			Preview:     t.PreviewURL,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	var req dto.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	input := service.CreateTripInput{
		Title:      req.Title,
		Location:   req.Location,
		StartDate:  utils.ParseDatePtr(req.StartDate),
		EndDate:    utils.ParseDatePtr(req.EndDate),
		PreviewURL: req.Preview,
		CreatedBy:  userID,
		IsPublic:   req.IsPublic,
	}
	trip, err := h.tripService.Create(r.Context(), input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateTripResponse{
		ID:      trip.ID,
		Preview: trip.PreviewURL,
	})
}

func (h *TripHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	log.Println("GetDetails called")

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		log.Println("Missing trip id")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing trip id"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.Printf("Invalid trip id: %s", idStr)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	log.Printf("Getting trip details for id=%d", id)

	trip, places, err := h.tripService.GetTripDetails(r.Context(), id)
	if err != nil {
		log.Printf("GetTripDetails error: %v", err)
		if err.Error() == "trip not found" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "trip not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	response := dto.TripDetailsResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Location:    trip.Location,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		Preview:     trip.PreviewURL,
		Attractions: places,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	var req dto.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	input := service.UpdateTripInput{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartDate:   utils.ParseDatePtr(req.StartDate),
		EndDate:     utils.ParseDatePtr(req.EndDate),
		PreviewURL:  req.Preview,
		IsPublic:    req.IsPublic,
	}
	_, err = h.tripService.Update(r.Context(), id, userID, input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	if err := h.tripService.Delete(r.Context(), id, userID); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TripHandler) GetTripPlaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}

	placeIDs, err := h.tripService.GetTripPlaceIDs(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch place IDs"})
		return
	}

	if placeIDs == nil {
		placeIDs = []uint64{}
	}

	json.NewEncoder(w).Encode(placeIDs)
}

func (h *TripHandler) AddPlace(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}

	var req struct {
		PlaceID    uint64 `json:"place_id"`
		OrderIndex int16  `json:"order_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if err := h.tripService.AddPlaceToTrip(r.Context(), tripID, req.PlaceID, userID, req.OrderIndex); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "place added to trip"})
}

func (h *TripHandler) RemovePlace(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}

	placeID, err := strconv.ParseUint(vars["placeId"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid place id"})
		return
	}

	if err := h.tripService.RemovePlaceFromTrip(r.Context(), tripID, placeID, userID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
