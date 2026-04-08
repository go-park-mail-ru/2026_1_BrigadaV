package handlers

import (
	"context"
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/models"
	"guidely-app/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TripService interface {
	Create(ctx context.Context, input service.CreateTripInput) (*models.Trip, error)
	GetByID(ctx context.Context, id uint64) (*models.Trip, error)
	GetUserTrips(ctx context.Context, userID uint64) ([]models.Trip, error)
	Update(ctx context.Context, id, userID uint64, input service.UpdateTripInput) (*models.Trip, error)
	Delete(ctx context.Context, id, userID uint64) error
}

type TripHandler struct {
	tripService TripService
}

func NewTripHandler(tripService TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

func parseDatePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var req dto.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	input := service.CreateTripInput{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   parseDatePtr(req.StartDate),
		EndDate:     parseDatePtr(req.EndDate),
		IsPublic:    req.IsPublic,
		CreatedBy:   userID,
	}
	trip, err := h.tripService.Create(r.Context(), input)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.TripResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Description: trip.Description,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		CreatedBy:   trip.CreatedBy,
		IsPublic:    trip.IsPublic,
		CreatedAt:   trip.CreatedAt,
		UpdatedAt:   trip.UpdatedAt,
	})
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	trips, err := h.tripService.GetUserTrips(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch trips"}`, http.StatusInternalServerError)
		return
	}
	resp := make([]dto.TripResponse, len(trips))
	for i, t := range trips {
		resp[i] = dto.TripResponse{
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
	json.NewEncoder(w).Encode(resp)
}

func (h *TripHandler) Get(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	id, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid trip id"}`, http.StatusBadRequest)
		return
	}
	trip, err := h.tripService.GetByID(r.Context(), id)
	if err != nil || trip == nil {
		http.Error(w, `{"error":"trip not found"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(dto.TripResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Description: trip.Description,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		CreatedBy:   trip.CreatedBy,
		IsPublic:    trip.IsPublic,
		CreatedAt:   trip.CreatedAt,
		UpdatedAt:   trip.UpdatedAt,
	})
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	id, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid trip id"}`, http.StatusBadRequest)
		return
	}
	var req dto.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	input := service.UpdateTripInput{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   parseDatePtr(req.StartDate),
		EndDate:     parseDatePtr(req.EndDate),
		IsPublic:    req.IsPublic != nil && *req.IsPublic,
	}
	trip, err := h.tripService.Update(r.Context(), id, userID, input)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(dto.TripResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Description: trip.Description,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		CreatedBy:   trip.CreatedBy,
		IsPublic:    trip.IsPublic,
		CreatedAt:   trip.CreatedAt,
		UpdatedAt:   trip.UpdatedAt,
	})
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	id, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid trip id"}`, http.StatusBadRequest)
		return
	}
	if err := h.tripService.Delete(r.Context(), id, userID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
