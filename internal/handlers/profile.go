package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"guidely-app/internal/storage"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type ProfileHandler struct {
	profileService service.ProfileService
	minioClient    *storage.MinioClient
}

func NewProfileHandler(profileService service.ProfileService, minioClient *storage.MinioClient) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		minioClient:    minioClient,
	}
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
		logger.Error(r.Context(), "GetProfile failed", logrus.Fields{"error": err, "user_id": userID})
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}
	response := dto.ProfileResponse{
		ID:         user.ID,
		Nickname:   user.Nickname,
		Login:      user.Login,
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
		logger.Error(r.Context(), "Invalid JSON in UpdateProfile", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	input := service.UpdateProfileInput{
		Nickname:  req.Nickname,
		Login:     req.Login,
		AvatarURL: req.AvatarURL,
		Country:   req.Country,
		City:      req.City,
		About:     req.About,
	}
	user, err := h.profileService.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		logger.Error(r.Context(), "UpdateProfile failed", logrus.Fields{"error": err, "user_id": userID})
		status := http.StatusBadRequest
		resp := map[string]string{"error": err.Error()}
		if err.Error() == "login already exists" {
			resp["field"] = "login"
		} else if err.Error() == "nickname already exists" {
			resp["field"] = "nickname"
		}
		if resp["field"] != "" {
			status = http.StatusConflict
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(resp)
		return
	}
	response := dto.ProfileResponse{
		ID:         user.ID,
		Nickname:   user.Nickname,
		Login:      user.Login,
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

func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	if h.minioClient == nil {
		logger.Error(r.Context(), "MinIO client not initialized", nil)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "storage not available"})
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		logger.Error(r.Context(), "Failed to parse form", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to parse form"})
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		logger.Error(r.Context(), "Missing avatar file", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing avatar file"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "file must be an image"})
		return
	}

	avatarURL, err := h.minioClient.UploadFile(r.Context(), file, header)
	if err != nil {
		logger.Error(r.Context(), "Failed to upload to MinIO", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to upload avatar"})
		return
	}

	logger.Info(r.Context(), "File uploaded to MinIO", logrus.Fields{"avatar_url": avatarURL})

	updatedUser, err := h.profileService.UpdateAvatar(r.Context(), userID, avatarURL)
	if err != nil {
		logger.Error(r.Context(), "UpdateAvatar failed", logrus.Fields{"error": err, "user_id": userID})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	response := dto.ProfileResponse{
		ID:         updatedUser.ID,
		Nickname:   updatedUser.Nickname,
		Login:      updatedUser.Login,
		AvatarURL:  updatedUser.AvatarURL,
		Country:    updatedUser.Country,
		City:       updatedUser.City,
		About:      updatedUser.About,
		HasReviews: updatedUser.HasReviews,
		CreatedAt:  updatedUser.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProfileHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.profileService.GetProfile(r.Context(), userID)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	if user.AvatarURL == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "avatar not set"})
		return
	}

	if strings.HasPrefix(user.AvatarURL, "http://") || strings.HasPrefix(user.AvatarURL, "https://") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"avatar_url": user.AvatarURL})
		return
	}

	filePath := strings.TrimPrefix(user.AvatarURL, "/uploads/")
	fullPath := filepath.Join("./uploads", filePath)

	ext := strings.ToLower(filepath.Ext(fullPath))
	contentType := "image/jpeg"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}
	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, fullPath)
}
