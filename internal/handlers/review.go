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

type ReviewService interface {
	Create(ctx context.Context, input service.CreateReviewInput) (*models.Review, error)
	GetByPlace(ctx context.Context, placeID uint64) ([]models.Review, error)
	Delete(ctx context.Context, userID, reviewID uint64) error
}

type ReviewHandler struct {
	reviewService ReviewService
}

func NewReviewHandler(reviewService ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	var visitDate *time.Time
	if req.VisitDate != nil && *req.VisitDate != "" {
		if t, err := time.Parse("2006-01-02", *req.VisitDate); err == nil {
			visitDate = &t
		}
	}
	input := service.CreateReviewInput{
		UserID:    userID,
		PlaceID:   req.PlaceID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		VisitDate: visitDate,
	}
	review, err := h.reviewService.Create(r.Context(), input)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.ReviewResponse{
		ID:        review.ID,
		UserID:    review.UserID,
		PlaceID:   review.PlaceID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		VisitDate: review.VisitDate,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	})
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	placeIDStr := r.URL.Query().Get("place_id")
	if placeIDStr == "" {
		http.Error(w, `{"error":"place_id required"}`, http.StatusBadRequest)
		return
	}
	placeID, err := strconv.ParseUint(placeIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid place_id"}`, http.StatusBadRequest)
		return
	}
	reviews, err := h.reviewService.GetByPlace(r.Context(), placeID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch reviews"}`, http.StatusInternalServerError)
		return
	}
	resp := make([]dto.ReviewResponse, len(reviews))
	for i, rv := range reviews {
		resp[i] = dto.ReviewResponse{
			ID:        rv.ID,
			UserID:    rv.UserID,
			PlaceID:   rv.PlaceID,
			Rating:    rv.Rating,
			Comment:   rv.Comment,
			VisitDate: rv.VisitDate,
			CreatedAt: rv.CreatedAt,
			UpdatedAt: rv.UpdatedAt,
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, `{"error":"invalid review id"}`, http.StatusBadRequest)
		return
	}
	if err := h.reviewService.Delete(r.Context(), userID, id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
