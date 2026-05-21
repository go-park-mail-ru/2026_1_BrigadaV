package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"guidely-app/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
		logger.Error(r.Context(), "Failed to fetch trips", logrus.Fields{"error": err})
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
		logger.Error(r.Context(), "Invalid JSON in CreateTrip", logrus.Fields{"error": err})
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
		logger.Error(r.Context(), "CreateTrip failed", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	logger.Info(r.Context(), "Trip created", logrus.Fields{"trip_id": trip.ID})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateTripResponse{
		ID:      trip.ID,
		Preview: trip.PreviewURL,
	})
}

func (h *TripHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing trip id"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in GetDetails", logrus.Fields{"id": idStr, "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}

	trip, places, err := h.tripService.GetTripDetails(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "GetTripDetails failed", logrus.Fields{"error": err, "trip_id": id})
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
		logger.Error(r.Context(), "Invalid trip id in Update", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	var req dto.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in UpdateTrip", logrus.Fields{"error": err})
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
		logger.Error(r.Context(), "UpdateTrip failed", logrus.Fields{"error": err, "trip_id": id})
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
		logger.Error(r.Context(), "trip delete error", logrus.Fields{"error": err, "trip_id": id})
		switch {
		case err.Error() == "trip not found":
			http.Error(w, "trip not found", http.StatusNotFound)
		case err.Error() == "not authorized":
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TripHandler) GetTripPlaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in GetTripPlaces", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}

	placeIDs, err := h.tripService.GetTripPlaceIDs(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch place IDs", logrus.Fields{"error": err, "trip_id": id})
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
		logger.Error(r.Context(), "Invalid trip id in AddPlace", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}

	var req struct {
		PlaceID    uint64 `json:"place_id"`
		OrderIndex int16  `json:"order_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in AddPlace", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if err := h.tripService.AddPlaceToTrip(r.Context(), tripID, req.PlaceID, userID, req.OrderIndex); err != nil {
		logger.Error(r.Context(), "AddPlaceToTrip failed", logrus.Fields{"error": err, "trip_id": tripID, "place_id": req.PlaceID})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	logger.Info(r.Context(), "Place added to trip", logrus.Fields{"trip_id": tripID, "place_id": req.PlaceID})
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
		logger.Error(r.Context(), "Invalid trip id in RemovePlace", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}

	placeID, err := strconv.ParseUint(vars["placeId"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid place id in RemovePlace", logrus.Fields{"placeId": vars["placeId"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid place id"})
		return
	}

	if err := h.tripService.RemovePlaceFromTrip(r.Context(), tripID, placeID, userID); err != nil {
		logger.Error(r.Context(), "RemovePlaceFromTrip failed", logrus.Fields{"error": err, "trip_id": tripID, "place_id": placeID})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	logger.Info(r.Context(), "Place removed from trip", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	w.WriteHeader(http.StatusNoContent)
}
