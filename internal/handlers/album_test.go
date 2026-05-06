package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "guidely-app/pkg/pb/album"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestAlbumHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient)

	reqBody := map[string]interface{}{
		"trip_id":     1,
		"name":        "Test",
		"description": "desc",
		"max_photos":  50,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/albums", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint64(1)))
	w := httptest.NewRecorder()

	mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&pb.Album{Id: 1, Name: "Test"}, nil)
	handler.Create(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAlbumHandler_Create_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient)

	req := httptest.NewRequest("POST", "/api/albums", nil)
	w := httptest.NewRecorder()

	handler.Create(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAlbumHandler_Create_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient)

	reqBody := map[string]interface{}{
		"trip_id": 1,
		"name":    "Test",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/albums", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint64(1)))
	w := httptest.NewRecorder()

	mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("internal"))
	handler.Create(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlbumHandler_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient)

	req := httptest.NewRequest("DELETE", "/api/albums/1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint64(1)))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().Delete(gomock.Any(), &pb.DeleteAlbumRequest{Id: 1}).Return(&emptypb.Empty{}, nil)
	handler.Delete(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAlbumHandler_Delete_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient)

	req := httptest.NewRequest("DELETE", "/api/albums/1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint64(1)))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	handler.Delete(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
