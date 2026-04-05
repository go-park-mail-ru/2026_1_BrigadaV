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
    userID, ok := r.Context().Value("user_id").(uint64)
    if !ok {
        http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    user, err := h.profileService.GetProfile(r.Context(), userID)
    if err != nil {
        http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
        return
    }
    
    json.NewEncoder(w).Encode(dto.ProfileResponse{
        ID:        user.ID,
        Login:     user.Login,
        Nickname:  user.Nickname,
        AvatarURL: user.AvatarURL,
        CreatedAt: user.CreatedAt,
    })
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value("user_id").(uint64)
    if !ok {
        http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    var req dto.UpdateProfileRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
        return
    }
    
    user, err := h.profileService.UpdateProfile(r.Context(), userID, req.Nickname, req.AvatarURL)
    if err != nil {
        http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
        return
    }
    
    json.NewEncoder(w).Encode(dto.ProfileResponse{
        ID:        user.ID,
        Login:     user.Login,
        Nickname:  user.Nickname,
        AvatarURL: user.AvatarURL,
        CreatedAt: user.CreatedAt,
    })
}

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value("user_id").(uint64)
    if !ok {
        http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
        return
    }
    
    var req dto.ChangePasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
        return
    }
    
    if err := h.profileService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
        http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{"message": "password changed successfully"})
}