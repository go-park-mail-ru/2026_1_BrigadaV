package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/internal/dto"
	"guidely-app/internal/service/mocks"
	"guidely-app/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestReviewHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewService := mocks.NewMockReviewService(ctrl)
	handler := NewReviewHandler(mockReviewService)

	reqBody := dto.CreateReviewRequest{
		PlaceID:   1,
		Title:     testutil.PtrString("Great"),
		Rating:    5,
		Content:   "Excellent place!",
		VisitDate: testutil.PtrString("2025-01-01"),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockReviewService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil)

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "ok", resp["message"])
}

func TestReviewHandler_Create_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewService := mocks.NewMockReviewService(ctrl)
	handler := NewReviewHandler(mockReviewService)

	reqBody := dto.CreateReviewRequest{PlaceID: 1, Rating: 5, Content: "Great"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReviewHandler_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewService := mocks.NewMockReviewService(ctrl)
	handler := NewReviewHandler(mockReviewService)

	req := httptest.NewRequest("DELETE", "/api/reviews/1", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockReviewService.EXPECT().Delete(gomock.Any(), uint64(1), uint64(1)).Return(nil)

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestReviewHandler_Delete_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewService := mocks.NewMockReviewService(ctrl)
	handler := NewReviewHandler(mockReviewService)

	req := httptest.NewRequest("DELETE", "/api/reviews/1", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReviewHandler_Delete_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewService := mocks.NewMockReviewService(ctrl)
	handler := NewReviewHandler(mockReviewService)

	req := httptest.NewRequest("DELETE", "/api/reviews/invalid", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReviewHandler_Delete_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewService := mocks.NewMockReviewService(ctrl)
	handler := NewReviewHandler(mockReviewService)

	req := httptest.NewRequest("DELETE", "/api/reviews/1", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockReviewService.EXPECT().Delete(gomock.Any(), uint64(1), uint64(1)).Return(errors.New("not authorized"))

	handler.Delete(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
