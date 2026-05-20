package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/internal/dto"
	"guidely-app/internal/repository/mocks"
	"guidely-app/internal/service"
	"guidely-app/internal/testutil"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestTripHandler_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	trips := []models.Trip{
		{ID: 1, Title: "Trip 1", Location: testutil.PtrString("Paris")},
		{ID: 2, Title: "Trip 2", Location: testutil.PtrString("London")},
	}

	req := httptest.NewRequest("GET", "/api/trips", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockTripRepo.EXPECT().GetByUser(gomock.Any(), uint64(1)).Return(trips, nil)

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []dto.TripResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "Trip 1", resp[0].Title)
}

func TestTripHandler_List_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	req := httptest.NewRequest("GET", "/api/trips", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTripHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	reqBody := dto.CreateTripRequest{
		Title:    "My Trip",
		Location: testutil.PtrString("Paris"),
		IsPublic: true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/trips", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockTripRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).DoAndReturn(func(_ context.Context, tr *models.Trip) error {
		tr.ID = 1
		return nil
	})

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.CreateTripResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, uint64(1), resp.ID)
}

func TestTripHandler_Create_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	reqBody := dto.CreateTripRequest{Title: "My Trip"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/trips", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTripHandler_GetDetails_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	trip := &models.Trip{ID: 1, Title: "My Trip", Location: testutil.PtrString("Paris")}
	places := []models.PlaceInTrip{{ID: 1, Name: "Eiffel Tower", Rating: 4.5}}

	req := httptest.NewRequest("GET", "/api/trips/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(trip, nil)
	mockTripRepo.EXPECT().GetAttractions(gomock.Any(), uint64(1)).Return(places, nil)

	handler.GetDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.TripDetailsResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, "My Trip", resp.Title)
	assert.Len(t, resp.Attractions, 1)
}

func TestTripHandler_GetDetails_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	req := httptest.NewRequest("GET", "/api/trips/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(999)).Return(nil, nil)

	handler.GetDetails(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTripHandler_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	reqBody := dto.UpdateTripRequest{Title: testutil.PtrString("Updated Title")}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/trips/1", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	trip := &models.Trip{ID: 1, Title: "Old Title", CreatedBy: 1}
	mockMemberRepo.EXPECT().HasEditPermission(gomock.Any(), uint64(1), uint64(1)).Return(true, nil)
	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(trip, nil)
	mockTripRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "ok", resp["message"])
}

func TestTripHandler_Update_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	reqBody := dto.UpdateTripRequest{Title: testutil.PtrString("Updated")}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/trips/1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTripHandler_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	req := httptest.NewRequest("DELETE", "/api/trips/1", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockMemberRepo.EXPECT().GetMemberRole(gomock.Any(), uint64(1), uint64(1)).Return("owner", nil)
	mockTripRepo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTripHandler_Delete_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	tripService := service.NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)
	handler := NewTripHandler(tripService)

	req := httptest.NewRequest("DELETE", "/api/trips/1", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
