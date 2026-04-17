package service

import (
	"context"
	"errors"
	"testing"

	"guidely-app/internal/models"
	"guidely-app/internal/repository/mocks"
	"guidely-app/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTripService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	input := CreateTripInput{
		Title:      "My Trip",
		Location:   testutil.PtrString("Paris"),
		PreviewURL: testutil.PtrString("/preview.jpg"),
		CreatedBy:  1,
		IsPublic:   true,
	}

	mockTripRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, tr *models.Trip) error {
		tr.ID = 1
		return nil
	})

	trip, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, "My Trip", trip.Title)
}

func TestTripService_Create_EmptyTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	input := CreateTripInput{
		Title:     "",
		CreatedBy: 1,
	}

	trip, err := svc.Create(context.Background(), input)
	assert.Error(t, err)
	assert.Equal(t, "title is required", err.Error())
	assert.Nil(t, trip)
}

func TestTripService_GetUserTrips_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	expectedTrips := []models.Trip{
		{ID: 1, Title: "Trip 1", CreatedBy: 1},
		{ID: 2, Title: "Trip 2", CreatedBy: 1},
	}

	mockTripRepo.EXPECT().GetByUser(gomock.Any(), uint64(1)).Return(expectedTrips, nil)

	trips, err := svc.GetUserTrips(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, trips, 2)
	assert.Equal(t, "Trip 1", trips[0].Title)
}

func TestTripService_GetUserTrips_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	mockTripRepo.EXPECT().GetByUser(gomock.Any(), uint64(1)).Return(nil, errors.New("db error"))

	trips, err := svc.GetUserTrips(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, trips)
}

func TestTripService_GetTripDetails_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	trip := &models.Trip{ID: 1, Title: "My Trip", CreatedBy: 1}
	places := []models.PlaceInTrip{{ID: 1, Name: "Eiffel Tower", Rating: 4.5}}

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(trip, nil)
	mockTripRepo.EXPECT().GetAttractions(gomock.Any(), uint64(1)).Return(places, nil)

	resultTrip, resultPlaces, err := svc.GetTripDetails(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, resultTrip)
	assert.Equal(t, "My Trip", resultTrip.Title)
	assert.Len(t, resultPlaces, 1)
	assert.Equal(t, "Eiffel Tower", resultPlaces[0].Name)
}

func TestTripService_GetTripDetails_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(nil, nil)

	trip, places, err := svc.GetTripDetails(context.Background(), 1)
	assert.Error(t, err)
	assert.Equal(t, "trip not found", err.Error())
	assert.Nil(t, trip)
	assert.Nil(t, places)
}

func TestTripService_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	existingTrip := &models.Trip{ID: 1, Title: "Old Title", CreatedBy: 1}
	input := UpdateTripInput{
		Title:       testutil.PtrString("New Title"),
		Description: testutil.PtrString("New description"),
		IsPublic:    testutil.PtrBool(false),
	}

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(existingTrip, nil)
	mockTripRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	updatedTrip, err := svc.Update(context.Background(), 1, 1, input)
	assert.NoError(t, err)
	assert.NotNil(t, updatedTrip)
	assert.Equal(t, "New Title", updatedTrip.Title)
}

func TestTripService_Update_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	existingTrip := &models.Trip{ID: 1, Title: "Trip", CreatedBy: 2}
	input := UpdateTripInput{Title: testutil.PtrString("New Title")}

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(existingTrip, nil)

	updatedTrip, err := svc.Update(context.Background(), 1, 1, input)
	assert.Error(t, err)
	assert.Equal(t, "not authorized", err.Error())
	assert.Nil(t, updatedTrip)
}

func TestTripService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	existingTrip := &models.Trip{ID: 1, CreatedBy: 1}

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(existingTrip, nil)
	mockTripRepo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.NoError(t, err)
}

func TestTripService_Delete_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	existingTrip := &models.Trip{ID: 1, CreatedBy: 2}

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(existingTrip, nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Equal(t, "not authorized", err.Error())
}

func TestTripService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	svc := NewTripService(mockTripRepo)

	mockTripRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(nil, nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Equal(t, "trip not found", err.Error())
}
