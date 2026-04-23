package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"guidely-app/internal/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
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
		logger.Error(r.Context(), "Invalid JSON in CreateReview", logrus.Fields{"error": err})
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
		VisitDate: utils.ParseDatePtr(req.VisitDate),
	}
	review, err := h.reviewService.Create(r.Context(), input)
	if err != nil {
		logger.Error(r.Context(), "CreateReview failed", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	logger.Info(r.Context(), "Review created", logrus.Fields{"review_id": review.ID})
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
		logger.Error(r.Context(), "Invalid review id in Delete", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid review id"})
		return
	}

	err = h.reviewService.Delete(r.Context(), userID, id)
	if err != nil {
		logger.Error(r.Context(), "DeleteReview failed", logrus.Fields{"error": err, "review_id": id})
		switch err.Error() {
		case "review not found":
			w.WriteHeader(http.StatusNotFound)
		case "not authorized":
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	logger.Info(r.Context(), "Review deleted", logrus.Fields{"review_id": id})
	w.WriteHeader(http.StatusNoContent)
}
