package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	input := service.CreateReviewInput{
		UserID:    userID,
		PlaceID:   req.PlaceID,
		Title:     req.Title,
		Rating:    req.Rating,
		Comment:   req.Content,
		VisitDate: parseDatePtr(req.VisitDate),
	}
	review, err := h.reviewService.Create(r.Context(), input)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id":      review.ID,
        "message": "ok",
    })
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid review id"})
		return
	}
	if err := h.reviewService.Delete(r.Context(), userID, id); err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
