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
	handler := NewAlbumHandler(mockClient, nil)

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
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("POST", "/api/albums", nil)
	w := httptest.NewRecorder()

	handler.Create(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAlbumHandler_Create_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	reqBody := map[string]interface{}{"trip_id": 1, "name": "Test"}
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
	handler := NewAlbumHandler(mockClient, nil)

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
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("DELETE", "/api/albums/1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint64(1)))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	handler.Delete(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlbumHandler_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("GET", "/api/albums/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().Get(gomock.Any(), &pb.GetAlbumRequest{Id: 1}).Return(&pb.Album{Id: 1, Name: "Test"}, nil)
	handler.Get(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlbumHandler_Get_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	handler := NewAlbumHandler(pb.NewMockAlbumServiceClient(ctrl), nil)

	req := httptest.NewRequest("GET", "/api/albums/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	handler.Get(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlbumHandler_Get_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("GET", "/api/albums/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
	handler.Get(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlbumHandler_GetPhotos_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("GET", "/api/albums/1/photos", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().GetPhotos(gomock.Any(), &pb.GetAlbumPhotosRequest{AlbumId: 1}).Return(&pb.GetAlbumPhotosResponse{
		Photos: []*pb.AlbumPhoto{{PhotoId: 1, FileUrl: "/photos/test.jpg"}},
	}, nil)
	handler.GetPhotos(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result []dto.PhotoUploadResponse
	json.NewDecoder(w.Body).Decode(&result)
	assert.Len(t, result, 1)
	assert.Equal(t, uint64(1), result[0].ID)
}

func TestAlbumHandler_GetPhotos_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("GET", "/api/albums/1/photos", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().GetPhotos(gomock.Any(), gomock.Any()).Return(nil, errors.New("grpc error"))
	handler.GetPhotos(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlbumHandler_RemovePhoto_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("DELETE", "/api/albums/1/photos/2", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1", "photoId": "2"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().RemovePhoto(gomock.Any(), &pb.RemovePhotoRequest{AlbumId: 1, PhotoId: 2}).Return(&emptypb.Empty{}, nil)
	handler.RemovePhoto(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAlbumHandler_RemovePhoto_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	handler := NewAlbumHandler(pb.NewMockAlbumServiceClient(ctrl), nil)

	req := httptest.NewRequest("DELETE", "/api/albums/abc/photos/2", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc", "photoId": "2"})
	w := httptest.NewRecorder()

	handler.RemovePhoto(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlbumHandler_GetByTrip_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("GET", "/api/trips/1/album", nil)
	req = mux.SetURLVars(req, map[string]string{"tripID": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().GetByTrip(gomock.Any(), &pb.GetAlbumByTripRequest{TripId: 1}).Return(&pb.Album{Id: 5, Name: "Основной альбом"}, nil)
	handler.GetByTrip(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlbumHandler_GetByTrip_AutoCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	req := httptest.NewRequest("GET", "/api/trips/1/album", nil)
	req = mux.SetURLVars(req, map[string]string{"tripID": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().GetByTrip(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
	mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&pb.Album{Id: 10, Name: "Основной альбом"}, nil)
	handler.GetByTrip(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlbumHandler_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := pb.NewMockAlbumServiceClient(ctrl)
	handler := NewAlbumHandler(mockClient, nil)

	reqBody := map[string]interface{}{"name": "Updated", "max_photos": 30}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/albums/1", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&pb.Album{Id: 1, Name: "Updated"}, nil)
	handler.Update(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
