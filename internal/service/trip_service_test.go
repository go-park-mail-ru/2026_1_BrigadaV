package service

import (
	"context"
	"testing"
	"time"

	"guidely-app/internal/repository/mocks"
	"guidely-app/internal/testutil"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTripService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

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
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

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
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

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

func TestTripService_GetTripDetails_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

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

func TestTripService_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	existingTrip := &models.Trip{ID: 1, Title: "Old Title", CreatedBy: 1}
	input := UpdateTripInput{
		Title:       testutil.PtrString("New Title"),
		Description: testutil.PtrString("New description"),
		IsPublic:    testutil.PtrBool(false),
	}

	mockMemberRepo.EXPECT().HasEditPermission(gomock.Any(), uint64(1), uint64(1)).Return(true, nil)
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
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	input := UpdateTripInput{Title: testutil.PtrString("New Title")}
	mockMemberRepo.EXPECT().HasEditPermission(gomock.Any(), uint64(1), uint64(1)).Return(false, nil)

	updatedTrip, err := svc.Update(context.Background(), 1, 1, input)
	assert.Error(t, err)
	assert.Equal(t, "not authorized to edit this trip", err.Error())
	assert.Nil(t, updatedTrip)
}

func TestTripService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockMemberRepo.EXPECT().GetMemberRole(gomock.Any(), uint64(1), uint64(1)).Return("owner", nil)
	mockTripRepo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.NoError(t, err)
}

func TestTripService_Delete_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockMemberRepo.EXPECT().GetMemberRole(gomock.Any(), uint64(1), uint64(1)).Return("viewer", nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Equal(t, "only owner can delete trip", err.Error())
}

func TestTripService_AddPlaceToTrip_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockMemberRepo.EXPECT().HasEditPermission(gomock.Any(), uint64(1), uint64(1)).Return(true, nil)
	mockTripRepo.EXPECT().CheckPlaceInTrip(gomock.Any(), uint64(1), uint64(5)).Return(false, nil)
	mockTripRepo.EXPECT().AddAttraction(gomock.Any(), uint64(1), uint64(5), int16(1)).Return(nil)
	err := svc.AddPlaceToTrip(context.Background(), 1, 5, 1, 1)
	assert.NoError(t, err)
}

func TestTripService_AddPlaceToTrip_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockMemberRepo.EXPECT().HasEditPermission(gomock.Any(), uint64(1), uint64(1)).Return(false, nil)
	err := svc.AddPlaceToTrip(context.Background(), 1, 5, 1, 1)
	assert.EqualError(t, err, "not authorized to edit this trip")
}

func TestTripService_RemovePlaceFromTrip_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockMemberRepo.EXPECT().HasEditPermission(gomock.Any(), uint64(1), uint64(1)).Return(true, nil)
	mockTripRepo.EXPECT().RemoveAttraction(gomock.Any(), uint64(1), uint64(5)).Return(nil)
	err := svc.RemovePlaceFromTrip(context.Background(), 1, 5, 1)
	assert.NoError(t, err)
}

func TestTripService_GetTripPlaceIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockTripRepo.EXPECT().GetPlaceIDs(gomock.Any(), uint64(1)).Return([]uint64{10, 20}, nil)
	ids, err := svc.GetTripPlaceIDs(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, []uint64{10, 20}, ids)
}

// Тесты для шеринга
func TestTripService_CreateViewShareLink_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockMemberRepo.EXPECT().GetMemberRole(gomock.Any(), uint64(1), uint64(100)).Return("owner", nil)
	mockInviteRepo.EXPECT().CreateInvite(gomock.Any(), gomock.Any()).Return(nil)

	link, err := svc.CreateViewShareLink(context.Background(), 1, 100)
	assert.NoError(t, err)
	assert.Contains(t, link, "/share/view/")
}

func TestTripService_CreateEditShareLink_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	mockMemberRepo.EXPECT().GetMemberRole(gomock.Any(), uint64(1), uint64(100)).Return("owner", nil)
	mockInviteRepo.EXPECT().CreateInvite(gomock.Any(), gomock.Any()).Return(nil)

	link, err := svc.CreateEditShareLink(context.Background(), 1, 100)
	assert.NoError(t, err)
	assert.Contains(t, link, "/share/edit/")
}

func TestTripService_AcceptInvite_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	invite := &models.TripInvite{ID: 1, TripID: 10, Role: "editor", IsOneTime: true}
	mockInviteRepo.EXPECT().GetInviteByToken(gomock.Any(), "token").Return(invite, nil)
	mockMemberRepo.EXPECT().AddMember(gomock.Any(), uint64(10), uint64(200), "editor").Return(nil)
	mockInviteRepo.EXPECT().MarkUsed(gomock.Any(), uint64(1)).Return(nil)

	tripID, role, err := svc.AcceptInvite(context.Background(), "token", 200)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), tripID)
	assert.Equal(t, "editor", role)
}

func TestTripService_AcceptInvite_Expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTripRepo := mocks.NewMockTripRepository(ctrl)
	mockMemberRepo := mocks.NewMockTripMemberRepository(ctrl)
	mockInviteRepo := mocks.NewMockTripInviteRepository(ctrl)
	svc := NewTripService(mockTripRepo, mockMemberRepo, mockInviteRepo)

	expired := testutil.PtrTime(time.Now().Add(-time.Hour))
	invite := &models.TripInvite{ID: 1, TripID: 10, Role: "editor", IsOneTime: true, ExpiresAt: expired}
	mockInviteRepo.EXPECT().GetInviteByToken(gomock.Any(), "token").Return(invite, nil)

	_, _, err := svc.AcceptInvite(context.Background(), "token", 200)
	assert.Error(t, err)
	assert.Equal(t, "invite has expired", err.Error())
}
