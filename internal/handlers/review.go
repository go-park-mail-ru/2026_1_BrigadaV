package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"guidely-app/internal/utils"
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
	input := service.CreateReviewInput{
		UserID:    userID,
		PlaceID:   req.PlaceID,
		Title:     req.Title,
		Rating:    req.Rating,
		Comment:   req.Content,
		VisitDate: utils.ParseDatePtr(req.VisitDate),
	}
	_, err = h.reviewService.Create(r.Context(), input)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusBadRequest)
		return
	}
	utils.WriteJSON(w, map[string]string{"message": "ok"}, http.StatusCreated)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	if err := h.reviewService.Delete(r.Context(), userID, id); err != nil {
		utils.WriteJSONError(w, err, http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
