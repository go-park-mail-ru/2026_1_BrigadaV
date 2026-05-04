package handlers

// import (
// 	"errors"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	pb "guidely-app/pkg/pb/album"

// 	"github.com/golang/mock/gomock"
// 	"github.com/gorilla/mux"
// 	"github.com/stretchr/testify/assert"
// )

// func TestAlbumHandler_Get_Error(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockClient := pb.NewMockAlbumServiceClient(ctrl)
// 	handler := NewAlbumHandler(mockClient)

// 	mockClient.EXPECT().Get(gomock.Any(), &pb.GetAlbumRequest{Id: 1}).Return(nil, errors.New("not found"))

// 	req := httptest.NewRequest("GET", "/api/albums/1", nil)
// 	req = mux.SetURLVars(req, map[string]string{"id": "1"})
// 	w := httptest.NewRecorder()

// 	handler.Get(w, req)

// 	assert.Equal(t, http.StatusInternalServerError, w.Code)
// }
