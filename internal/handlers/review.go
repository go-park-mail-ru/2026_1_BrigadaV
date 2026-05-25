package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	pb "guidely-app/pkg/pb/review"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
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
		w.Header().Set("Content-Type", "application/json")

		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				logger.Warn(r.Context(), "duplicate review attempt", logrus.Fields{
					"user_id":  userID,
					"place_id": req.PlaceID,
				})
				w.WriteHeader(http.StatusConflict) // 409
				json.NewEncoder(w).Encode(map[string]string{
					"error": "you have already reviewed this place",
				})
				return
			}
		}

		logger.Error(r.Context(), "review create error", logrus.Fields{
			"error":    err,
			"user_id":  userID,
			"place_id": req.PlaceID,
		})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}

	logger.Info(r.Context(), "review created", logrus.Fields{
		"review_id": resp.Id,
		"user_id":   userID,
		"place_id":  req.PlaceID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      resp.Id,
		"message": "ok",
	})
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	vars := mux.Vars(r)
	reviewID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid review id"})
		return
	}

	_, err = h.client.DeleteReview(r.Context(), &pb.DeleteReviewRequest{
		UserId:   userID,
		ReviewId: reviewID,
	})
	if err != nil {
		logger.Error(r.Context(), "review delete error", logrus.Fields{
			"error":     err,
			"review_id": reviewID,
			"user_id":   userID,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
