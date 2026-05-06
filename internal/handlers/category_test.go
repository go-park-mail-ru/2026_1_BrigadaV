package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/internal/repository/mocks"
	"guidely-app/internal/service"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCategoryHandler_List_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockCategoryRepository(ctrl)
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	mockRepo.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))
	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCategoryHandler_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockCategoryRepository(ctrl)
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	categories := []models.Category{{ID: 1, Name: "Hotel", ApplicableTypes: []string{"hotel"}}}
	mockRepo.EXPECT().GetAll(gomock.Any()).Return(categories, nil)
	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.Category
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
}

func TestCategoryHandler_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockCategoryRepository(ctrl)
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	cat := &models.Category{ID: 1, Name: "Hotel"}
	mockRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(cat, nil)
	req := httptest.NewRequest("GET", "/api/categories/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()
	handler.Get(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCategoryHandler_Get_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockCategoryRepository(ctrl)
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	mockRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(nil, nil)
	req := httptest.NewRequest("GET", "/api/categories/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()
	handler.Get(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoryHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockCategoryRepository(ctrl)
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	body := `{"name":"New","description":"desc","applicable_types":["attraction"]}`
	req := httptest.NewRequest("POST", "/api/categories", bytes.NewReader([]byte(body)))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	handler.Create(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCategoryHandler_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockCategoryRepository(ctrl)
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	body := `{"name":"Updated","description":"new","applicable_types":["hotel"]}`
	req := httptest.NewRequest("PUT", "/api/categories/1", bytes.NewReader([]byte(body)))
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	handler.Update(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCategoryHandler_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockCategoryRepository(ctrl)
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	req := httptest.NewRequest("DELETE", "/api/categories/1", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	mockRepo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)
	handler.Delete(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
