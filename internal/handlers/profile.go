package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	"guidely-app/internal/service"
	"guidely-app/pkg/storage"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
)

const avatarUploadDir = "./uploads/avatars"

type ProfileHandler struct {
	profileService service.ProfileService
	s3             *storage.S3Client
}

func NewProfileHandler(profileService service.ProfileService, s3 *storage.S3Client) *ProfileHandler {
	return &ProfileHandler{profileService: profileService, s3: s3}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	user, err := h.profileService.GetProfile(r.Context(), userID)
	if err != nil {
		logger.Error(r.Context(), "GetProfile failed", logrus.Fields{"error": err, "user_id": userID})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"user not found"}`))
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
	data, err := easyjson.Marshal(&response)
	if err == nil {
		w.Write(data)
	}
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}

	var req dto.UpdateProfileRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in UpdateProfile", logrus.Fields{"error": err})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid request body"}`))
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
		w.Header().Set("Content-Type", "application/json")
		if err.Error() == "nickname already taken" {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error":"nickname already taken"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
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
	data, err := easyjson.Marshal(&response)
	if err == nil {
		w.Write(data)
	}
}

// UploadAvatar – загружает аватар: в S3 (если включён) или локально
func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}

	const maxAvatarSize = 5 << 20 // 5 МБ
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		logger.Error(r.Context(), "ParseMultipartForm failed", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"file too large or invalid form"}`))
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		logger.Error(r.Context(), "Missing avatar file", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"missing avatar file"}`))
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"file too large (max 5 MB)"}`))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"file must be an image"}`))
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	var avatarURL string

	if h.s3 != nil {
		avatarURL, err = h.s3.UploadFile(r.Context(), "avatars/"+objectName, file, header.Size, contentType)
		if err != nil {
			logger.Error(r.Context(), "S3 upload failed", logrus.Fields{"error": err, "user_id": userID})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"failed to upload avatar"}`))
			return
		}
		logger.Info(r.Context(), "Avatar uploaded to S3", logrus.Fields{"url": avatarURL, "user_id": userID})
	} else {
		if err := os.MkdirAll(avatarUploadDir, 0o755); err != nil {
			logger.Error(r.Context(), "mkdir error", logrus.Fields{"error": err})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal error"}`))
			return
		}
		localPath := filepath.Join(avatarUploadDir, objectName)
		dst, err := os.Create(localPath)
		if err != nil {
			logger.Error(r.Context(), "file create error", logrus.Fields{"error": err, "path": localPath})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal error"}`))
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			logger.Error(r.Context(), "file copy error", logrus.Fields{"error": err})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal error"}`))
			return
		}
		avatarURL = "/uploads/avatars/" + objectName
		logger.Info(r.Context(), "Avatar saved locally", logrus.Fields{"path": avatarURL, "user_id": userID})
	}

	updatedUser, err := h.profileService.UpdateAvatar(r.Context(), userID, avatarURL)
	if err != nil {
		logger.Error(r.Context(), "UpdateAvatar failed", logrus.Fields{"error": err, "user_id": userID})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
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
	data, err := easyjson.Marshal(&response)
	if err == nil {
		w.Write(data)
	}
}

func (h *ProfileHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}

	user, err := h.profileService.GetProfile(r.Context(), userID)
	if err != nil || user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"user not found"}`))
		return
	}

	if user.AvatarURL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"avatar not set"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"avatar_url":"` + user.AvatarURL + `"}`))
}
