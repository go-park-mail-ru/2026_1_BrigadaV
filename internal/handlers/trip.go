package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"guidely-app/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type TripHandler struct {
	tripService *service.TripService
}

func NewTripHandler(tripService *service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

func parseDatePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.DateOnly, *s)
	if err != nil {
		return nil
	}
	return &t
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	trips, err := h.tripService.GetUserTrips(r.Context(), userID)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		return
	}
	response := make([]dto.TripResponse, len(trips))
	for i, t := range trips {
		response[i] = dto.TripResponse{
			ID:        t.ID,
			Title:     t.Title,
			Location:  t.Location,
			StartDate: t.StartDate,
			EndDate:   t.EndDate,
			Preview:   t.PreviewURL,
		}
	}
	utils.WriteJSON(w, response, http.StatusOK)
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	var req dto.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	input := service.CreateTripInput{
		Title:      req.Title,
		Location:   req.Location,
		StartDate:  parseDatePtr(req.StartDate),
		EndDate:    parseDatePtr(req.EndDate),
		PreviewURL: req.Preview,
		CreatedBy:  userID,
		IsPublic:   req.IsPublic,
	}
	trip, err := h.tripService.Create(r.Context(), input)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusBadRequest)
		return
	}
	utils.WriteJSON(w, dto.CreateTripResponse{
		ID:      trip.ID,
		Preview: trip.PreviewURL,
	}, http.StatusCreated)
}

func (h *TripHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	trip, places, err := h.tripService.GetTripDetails(r.Context(), id)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrNotFound, http.StatusNotFound)
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
	utils.WriteJSON(w, response, http.StatusOK)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	var req dto.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	input := service.UpdateTripInput{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartDate:   parseDatePtr(req.StartDate),
		EndDate:     parseDatePtr(req.EndDate),
		PreviewURL:  req.Preview,
		IsPublic:    req.IsPublic,
	}
	_, err = h.tripService.Update(r.Context(), id, userID, input)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusBadRequest)
		return
	}
	utils.WriteJSON(w, map[string]string{"message": "ok"}, http.StatusOK)
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	if err := h.tripService.Delete(r.Context(), id, userID); err != nil {
		utils.WriteJSONError(w, err, http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
