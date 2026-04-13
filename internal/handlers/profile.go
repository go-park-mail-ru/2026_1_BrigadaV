package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"net/http"
)

type ProfileHandler struct {
	profileService *service.ProfileService
}

func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	user, err := h.profileService.GetProfile(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
