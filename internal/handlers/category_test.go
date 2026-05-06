package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/internal/repository/mocks"
	"guidely-app/internal/service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCategoryHandler_List_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCategoryRepository(ctrl) // мок появится после генерации
	svc := service.NewCategoryService(mockRepo)
	handler := NewCategoryHandler(svc)

	mockRepo.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal error")
}
