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
		Title:       req.Title,
		Description: req.Description,
		StartDate:   parseDatePtr(req.StartDate),
		EndDate:     parseDatePtr(req.EndDate),
		CreatedBy:   userID,
		IsPublic:    req.IsPublic,
	}
	trip, err := h.tripService.Create(r.Context(), input)
	if err != nil {
		switch err {
		case utils.ErrBadRequest:
			utils.WriteJSONError(w, err, http.StatusBadRequest)
		default:
			utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, dto.TripResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Description: trip.Description,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		CreatedBy:   trip.CreatedBy,
		IsPublic:    trip.IsPublic,
		CreatedAt:   trip.CreatedAt,
		UpdatedAt:   trip.UpdatedAt,
	}, http.StatusCreated)
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
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			StartDate:   t.StartDate,
			EndDate:     t.EndDate,
			CreatedBy:   t.CreatedBy,
			IsPublic:    t.IsPublic,
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		}
	}
	utils.WriteJSON(w, response, http.StatusOK)
}

func (h *TripHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	trip, err := h.tripService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrNotFound, http.StatusNotFound)
		return
	}
	utils.WriteJSON(w, dto.TripResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Description: trip.Description,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		CreatedBy:   trip.CreatedBy,
		IsPublic:    trip.IsPublic,
		CreatedAt:   trip.CreatedAt,
		UpdatedAt:   trip.UpdatedAt,
	}, http.StatusOK)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
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
		StartDate:   parseDatePtr(req.StartDate),
		EndDate:     parseDatePtr(req.EndDate),
		IsPublic:    req.IsPublic,
	}
	trip, err := h.tripService.Update(r.Context(), id, userID, input)
	if err != nil {
		switch err {
		case utils.ErrNotFound:
			utils.WriteJSONError(w, err, http.StatusNotFound)
		case utils.ErrUnauthorized:
			utils.WriteJSONError(w, err, http.StatusForbidden)
		case utils.ErrBadRequest:
			utils.WriteJSONError(w, err, http.StatusBadRequest)
		default:
			utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, dto.TripResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Description: trip.Description,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		CreatedBy:   trip.CreatedBy,
		IsPublic:    trip.IsPublic,
		CreatedAt:   trip.CreatedAt,
		UpdatedAt:   trip.UpdatedAt,
	}, http.StatusOK)
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	if err := h.tripService.Delete(r.Context(), id, userID); err != nil {
		switch err {
		case utils.ErrNotFound:
			utils.WriteJSONError(w, err, http.StatusNotFound)
		case utils.ErrUnauthorized:
			utils.WriteJSONError(w, err, http.StatusForbidden)
		default:
			utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
