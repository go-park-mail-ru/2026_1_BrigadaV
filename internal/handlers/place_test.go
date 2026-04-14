package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/internal/dto"
	"guidely-app/internal/models"
	"guidely-app/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPlaceHandler_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceService := mocks.NewMockPlaceService(ctrl)
	handler := NewPlaceHandler(mockPlaceService)

	places := []models.Place{
		{ID: 1, Name: "Place 1", Description: "Desc 1", Price: 1000},
		{ID: 2, Name: "Place 2", Description: "Desc 2", Price: 2000},
	}

	mockPlaceService.EXPECT().GetAll(gomock.Any()).Return(places, nil)

	req := httptest.NewRequest("GET", "/api/places", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []dto.PlaceResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "Place 1", resp[0].Name)
}

func TestPlaceHandler_List_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceService := mocks.NewMockPlaceService(ctrl)
	handler := NewPlaceHandler(mockPlaceService)

	mockPlaceService.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/api/places", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPlaceHandler_GetDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceService := mocks.NewMockPlaceService(ctrl)
	handler := NewPlaceHandler(mockPlaceService)

	place := &models.PlaceWithRating{
		ID:          1,
		Name:        "Eiffel Tower",
		Description: "Famous tower",
		Price:       1500,
		Rating:      4.8,
		ReviewCount: 100,
		IsLiked:     false,
	}

	mockPlaceService.EXPECT().GetDetails(gomock.Any(), uint64(1), uint64(0)).Return(place, nil)

	req := httptest.NewRequest("GET", "/api/places/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.GetDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.PlaceWithRating
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, "Eiffel Tower", resp.Name)
}

func TestPlaceHandler_GetDetails_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceService := mocks.NewMockPlaceService(ctrl)
	handler := NewPlaceHandler(mockPlaceService)

	mockPlaceService.EXPECT().GetDetails(gomock.Any(), uint64(999), uint64(0)).Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/api/places/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	handler.GetDetails(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPlaceHandler_GetReviews(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceService := mocks.NewMockPlaceService(ctrl)
	handler := NewPlaceHandler(mockPlaceService)

	reviews := []models.ReviewWithAuthor{
		{ID: 1, Rating: 5, Comment: "Great!", Author: struct {
			ID       uint64  `json:"id"`
			Nickname string  `json:"nickname"`
			Avatar   *string `json:"avatar,omitempty"`
		}{ID: 1, Nickname: "john"}},
	}

	mockPlaceService.EXPECT().GetReviews(gomock.Any(), uint64(1)).Return(reviews, nil)

	req := httptest.NewRequest("GET", "/api/places/1/reviews", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.GetReviews(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []models.ReviewWithAuthor
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Great!", resp[0].Comment)
}
