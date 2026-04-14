package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"guidely-app/internal/dto"
	"guidely-app/internal/models"
	"guidely-app/internal/service/mocks"
	"guidely-app/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestProfileHandler_GetProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileService := mocks.NewMockProfileService(ctrl)
	handler := NewProfileHandler(mockProfileService)

	user := &models.User{
		ID:         1,
		Nickname:   "johnny",
		AvatarURL:  "/avatar.jpg",
		Country:    testutil.PtrString("USA"),
		City:       testutil.PtrString("NYC"),
		About:      testutil.PtrString("About me"),
		HasReviews: true,
		CreatedAt:  time.Now(),
	}

	req := httptest.NewRequest("GET", "/api/profile", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockProfileService.EXPECT().GetProfile(gomock.Any(), uint64(1)).Return(user, nil)

	handler.GetProfile(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ProfileResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, "johnny", resp.Nickname)
}

func TestProfileHandler_GetProfile_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileService := mocks.NewMockProfileService(ctrl)
	handler := NewProfileHandler(mockProfileService)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	w := httptest.NewRecorder()

	handler.GetProfile(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProfileHandler_UpdateProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileService := mocks.NewMockProfileService(ctrl)
	handler := NewProfileHandler(mockProfileService)

	reqBody := dto.UpdateProfileRequest{
		Nickname:  testutil.PtrString("new_nick"),
		AvatarURL: testutil.PtrString("/new_avatar.jpg"),
		Country:   testutil.PtrString("Canada"),
		City:      testutil.PtrString("Toronto"),
		About:     testutil.PtrString("New about"),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/profile", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	updatedUser := &models.User{
		ID:        1,
		Nickname:  "new_nick",
		AvatarURL: "/new_avatar.jpg",
		Country:   testutil.PtrString("Canada"),
		City:      testutil.PtrString("Toronto"),
		About:     testutil.PtrString("New about"),
	}

	mockProfileService.EXPECT().UpdateProfile(gomock.Any(), uint64(1), gomock.Any()).Return(updatedUser, nil)

	handler.UpdateProfile(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ProfileResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "new_nick", resp.Nickname)
}

func TestProfileHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileService := mocks.NewMockProfileService(ctrl)
	handler := NewProfileHandler(mockProfileService)

	req := httptest.NewRequest("PUT", "/api/profile", bytes.NewReader([]byte(`{invalid json}`)))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UpdateProfile(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_UpdateProfile_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileService := mocks.NewMockProfileService(ctrl)
	handler := NewProfileHandler(mockProfileService)

	reqBody := dto.UpdateProfileRequest{Nickname: testutil.PtrString("new")}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/profile", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateProfile(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
