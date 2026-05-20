package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"guidely-app/pkg/storage"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ProfileHandler struct {
	profileService service.ProfileService
	s3             *storage.S3Client
}

func NewProfileHandler(profileService service.ProfileService, s3 *storage.S3Client) *ProfileHandler {
	return &ProfileHandler{profileService: profileService, s3: s3}
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
		AvatarURL: req.AvatarURL,
		Country:   req.Country,
		City:      req.City,
		About:     req.About,
	}
	user, err := h.profileService.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		logger.Error(r.Context(), "UpdateProfile failed", logrus.Fields{"error": err, "user_id": userID})
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

// UploadAvatar принимает multipart/form-data с полем "avatar",
// загружает файл в S3/MinIO и обновляет avatar_url пользователя.
func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Сначала проверяем авторизацию
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	// Лимит 10 МБ на весь multipart
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logger.Error(r.Context(), "ParseMultipartForm failed", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "request too large or malformed"})
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

	// Проверяем доступность S3 только после того, как файл получен и валиден
	if h.s3 == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "avatar upload is disabled (S3 not configured)"})
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	// Уникальное имя объекта: avatars/<uuid>.jpg
	objectName := fmt.Sprintf("avatars/%s%s", uuid.New().String(), ext)

	avatarURL, err := h.s3.UploadFile(r.Context(), objectName, file, header.Size, contentType)
	if err != nil {
		logger.Error(r.Context(), "S3 upload failed", logrus.Fields{"error": err, "user_id": userID})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to upload avatar"})
		return
	}

	logger.Info(r.Context(), "Avatar uploaded to S3", logrus.Fields{
		"url":     avatarURL,
		"user_id": userID,
	})

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

// GetAvatar — возвращает URL аватара
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"avatar_url": user.AvatarURL})
}
