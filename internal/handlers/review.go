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

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	var visitDate *time.Time
	if req.VisitDate != nil && *req.VisitDate != "" {
		t, err := time.Parse(time.DateOnly, *req.VisitDate)
		if err == nil {
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
		switch err {
		case utils.ErrBadRequest:
			utils.WriteJSONError(w, err, http.StatusBadRequest)
		default:
			utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, dto.ReviewResponse{
		ID:        review.ID,
		UserID:    review.UserID,
		PlaceID:   review.PlaceID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		VisitDate: review.VisitDate,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}, http.StatusCreated)
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	placeIDStr := r.URL.Query().Get("place_id")
	if placeIDStr == "" {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	placeID, err := strconv.ParseUint(placeIDStr, 10, 64)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	reviews, err := h.reviewService.GetByPlace(r.Context(), placeID)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		return
	}
	response := make([]dto.ReviewResponse, len(reviews))
	for i, rv := range reviews {
		response[i] = dto.ReviewResponse{
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
	utils.WriteJSON(w, response, http.StatusOK)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	if err := h.reviewService.Delete(r.Context(), userID, id); err != nil {
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
