package album

import (
	"context"
	"testing"

	"guidely-app/internal/album/repository/mocks"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAlbumRepository(ctrl)
	svc := NewService(repo)

	album := &models.Album{TripID: 1, Name: "Test"}
	repo.EXPECT().Create(gomock.Any(), album).Return(nil)

	err := svc.Create(context.Background(), album)
	assert.NoError(t, err)
}
