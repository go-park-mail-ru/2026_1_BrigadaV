package handlers

import (
	"net/http"
	"strconv"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	pb "guidely-app/pkg/pb/review"

	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
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
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.CreateReviewRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid request"})
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
				writeJSON(w, http.StatusConflict, &dto.ErrorResponse{Error: "you have already reviewed this place"})
				return
			}
		}
		logger.Error(r.Context(), "review create error", logrus.Fields{"error": err, "user_id": userID, "place_id": req.PlaceID})
		writeJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Error: "internal error"})
		return
	}

	logger.Info(r.Context(), "review created", logrus.Fields{
		"review_id": resp.Id,
		"user_id":   userID,
		"place_id":  req.PlaceID,
	})

	writeJSON(w, http.StatusCreated, &dto.ReviewCreatedResponse{ID: resp.Id, Message: "ok"})
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	vars := mux.Vars(r)
	reviewID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid review id"})
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
		writeJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Error: "internal error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
