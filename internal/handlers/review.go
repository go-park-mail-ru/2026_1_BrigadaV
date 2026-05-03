package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"guidely-app/internal/middleware"
	pb "guidely-app/pkg/pb/review"

	"github.com/gorilla/mux"
)

type ReviewHandler struct {
	client pb.ReviewServiceClient
}

func NewReviewHandler(client pb.ReviewServiceClient) *ReviewHandler {
	return &ReviewHandler{client: client}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		PlaceID   uint64  `json:"place_id"`
		Title     *string `json:"title"`
		Rating    int16   `json:"rating"`
		Content   string  `json:"content"`
		VisitDate *string `json:"visit_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateReview(r.Context(), &pb.CreateReviewRequest{
		UserId:    userID,
		PlaceId:   req.PlaceID,
		Title:     req.Title,
		Rating:    int32(req.Rating),
		Comment:   req.Content,
		VisitDate: req.VisitDate,
	})
	if err != nil {
		log.Printf("review create error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      resp.Id,
		"message": "ok",
	})
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	reviewID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid review id", http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteReview(r.Context(), &pb.DeleteReviewRequest{
		UserId:   userID,
		ReviewId: reviewID,
	})
	if err != nil {
		log.Printf("review delete error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
