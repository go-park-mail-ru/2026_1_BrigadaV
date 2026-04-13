package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"guidely-app/internal/utils"
	"net/http"
)

type ProfileHandler struct {
	profileService *service.ProfileService
}

func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	user, err := h.profileService.GetProfile(r.Context(), userID)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrNotFound, http.StatusNotFound)
		return
	}
	response := dto.ProfileResponse{
		ID:         user.ID,
		Nickname:   user.Nickname,
		AvatarURL:  user.AvatarURL,
		Country:    user.Country,
		City:       user.City,
		About:      user.About,
		HasReviews: user.HasReviews,
		CreatedAt:  user.CreatedAt,
	}
	utils.WriteJSON(w, response, http.StatusOK)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	input := service.UpdateProfileInput{
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
		Country:   req.Country,
		City:      req.City,
		About:     req.About,
	}
	user, err := h.profileService.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusBadRequest)
		return
	}
	response := dto.ProfileResponse{
		ID:         user.ID,
		Nickname:   user.Nickname,
		AvatarURL:  user.AvatarURL,
		Country:    user.Country,
		City:       user.City,
		About:      user.About,
		HasReviews: user.HasReviews,
		CreatedAt:  user.CreatedAt,
	}
	utils.WriteJSON(w, response, http.StatusOK)
}
