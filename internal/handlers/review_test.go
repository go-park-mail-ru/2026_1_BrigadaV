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
	pb "guidely-app/pkg/pb/review"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestReviewHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	reqBody := dto.CreateReviewRequest{
		PlaceID: 1,
		Rating:  5,
		Content: "Excellent place!",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockClient.EXPECT().CreateReview(gomock.Any(), gomock.Any()).Return(&pb.ReviewResponse{Id: 1}, nil)

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp dto.ReviewCreatedResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "ok", resp.Message)
	assert.Equal(t, uint64(1), resp.ID)
}

func TestReviewHandler_Create_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	reqBody := dto.CreateReviewRequest{PlaceID: 1, Rating: 5, Content: "Great"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReviewHandler_Create_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	reqBody := dto.CreateReviewRequest{PlaceID: 1, Rating: 5, Content: "Great"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockClient.EXPECT().CreateReview(gomock.Any(), gomock.Any()).Return(nil, errors.New("grpc error"))

	handler.Create(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestReviewHandler_Create_AlreadyExists проверяет что при попытке создать второй отзыв
// к одной достопримечательности возвращается 409 Conflict с понятным сообщением.
func TestReviewHandler_Create_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	reqBody := dto.CreateReviewRequest{PlaceID: 1, Rating: 4, Content: "Second review"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockClient.EXPECT().CreateReview(gomock.Any(), gomock.Any()).Return(
		nil,
		status.Error(codes.AlreadyExists, "you have already reviewed this place"),
	)

	handler.Create(w, req)

	assert.Equal(t, http.StatusConflict, w.Code) // 409
	var resp dto.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp.Error, "already reviewed")
}

func TestReviewHandler_Create_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader([]byte("not json")))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Create(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReviewHandler_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	req := httptest.NewRequest("DELETE", "/api/reviews/1", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().DeleteReview(gomock.Any(), &pb.DeleteReviewRequest{UserId: 1, ReviewId: 1}).Return(nil, nil)

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestReviewHandler_Delete_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	req := httptest.NewRequest("DELETE", "/api/reviews/1", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReviewHandler_Delete_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	req := httptest.NewRequest("DELETE", "/api/reviews/invalid", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReviewHandler_Delete_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := pb.NewMockReviewServiceClient(ctrl)
	handler := NewReviewHandler(mockClient)

	req := httptest.NewRequest("DELETE", "/api/reviews/1", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockClient.EXPECT().DeleteReview(gomock.Any(), gomock.Any()).Return(nil, errors.New("grpc error"))

	handler.Delete(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
