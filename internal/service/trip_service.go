package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"time"

	"guidely-app/internal/logger"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"

	"github.com/sirupsen/logrus"
)

// Определения типов входных данных
type CreateTripInput struct {
	Title      string
	Location   *string
	StartDate  *time.Time
	EndDate    *time.Time
	PreviewURL *string
	CreatedBy  uint64
	IsPublic   bool
}

type UpdateTripInput struct {
	Title       *string
	Description *string
	Location    *string
	StartDate   *time.Time
	EndDate     *time.Time
	PreviewURL  *string
	IsPublic    *bool
}

type tripService struct {
	tripRepo   repository.TripRepository
	memberRepo repository.TripMemberRepository
	inviteRepo repository.TripInviteRepository
}

func NewTripService(
	tripRepo repository.TripRepository,
	memberRepo repository.TripMemberRepository,
	inviteRepo repository.TripInviteRepository,
) TripService {
	return &tripService{
		tripRepo:   tripRepo,
		memberRepo: memberRepo,
		inviteRepo: inviteRepo,
	}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func getShareBaseURL() string {
	if base := os.Getenv("SHARE_BASE_URL"); base != "" {
		return base
	}
	return "http://localhost:8080"
}

// Create – создание новой поездки (владелец автоматически добавляется в trip_member через триггер)
func (s *tripService) Create(ctx context.Context, input CreateTripInput) (*models.Trip, error) {
	logger.Info(ctx, "CreateTrip called", logrus.Fields{"title": input.Title})
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	trip := &models.Trip{
		Title:      input.Title,
		Location:   input.Location,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		PreviewURL: input.PreviewURL,
		CreatedBy:  input.CreatedBy,
		IsPublic:   input.IsPublic,
	}
	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}
	// После создания триггер добавит запись в trip_member с ролью owner
	logger.Info(ctx, "CreateTrip successful", logrus.Fields{"trip_id": trip.ID})
	return trip, nil
}

// GetUserTrips – поездки, где пользователь является участником (owner/editor/viewer)
func (s *tripService) GetUserTrips(ctx context.Context, userID uint64) ([]models.Trip, error) {
	// Получаем все trip_id из trip_member для этого userID
	// Реализуем новый метод в репозитории или используем существующий GetByUser (который возвращает только созданные)
	// Для совместных поездок нужно отдельное получение. Пока оставим как есть (только created_by).
	// В идеале добавить метод GetTripsByMember.
	return s.tripRepo.GetByUser(ctx, userID)
}

// GetTripDetails – получение поездки и её достопримечательностей (без проверки прав – вызывать после авторизации)
func (s *tripService) GetTripDetails(ctx context.Context, tripID uint64) (*models.Trip, []models.PlaceInTrip, error) {
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, nil, err
	}
	if trip == nil {
		return nil, nil, errors.New("trip not found")
	}
	places, err := s.tripRepo.GetAttractions(ctx, tripID)
	if err != nil {
		return nil, nil, err
	}
	return trip, places, nil
}

// Update – обновление поездки с проверкой прав (владелец или редактор)
func (s *tripService) Update(ctx context.Context, id, userID uint64, input UpdateTripInput) (*models.Trip, error) {
	logger.Info(ctx, "UpdateTrip called", logrus.Fields{"trip_id": id, "user_id": userID})
	ok, err := s.memberRepo.HasEditPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not authorized to edit this trip")
	}
	trip, err := s.tripRepo.GetByID(ctx, id)
	if err != nil || trip == nil {
		return nil, errors.New("trip not found")
	}
	if input.Title != nil {
		trip.Title = *input.Title
	}
	if input.Description != nil {
		trip.Description = *input.Description
	}
	if input.Location != nil {
		trip.Location = input.Location
	}
	if input.StartDate != nil {
		trip.StartDate = input.StartDate
	}
	if input.EndDate != nil {
		trip.EndDate = input.EndDate
	}
	if input.PreviewURL != nil {
		trip.PreviewURL = input.PreviewURL
	}
	if input.IsPublic != nil {
		trip.IsPublic = *input.IsPublic
	}
	if err := s.tripRepo.Update(ctx, trip); err != nil {
		return nil, err
	}
	logger.Info(ctx, "UpdateTrip successful", logrus.Fields{"trip_id": trip.ID})
	return trip, nil
}

// Delete – удаление поездки (только владелец)
func (s *tripService) Delete(ctx context.Context, id, userID uint64) error {
	logger.Info(ctx, "DeleteTrip called", logrus.Fields{"trip_id": id, "user_id": userID})
	role, err := s.memberRepo.GetMemberRole(ctx, id, userID)
	if err != nil {
		return err
	}
	if role != "owner" {
		return errors.New("only owner can delete trip")
	}
	return s.tripRepo.Delete(ctx, id)
}

// GetTripPlaceIDs – список ID достопримечательностей (без проверки прав)
func (s *tripService) GetTripPlaceIDs(ctx context.Context, tripID uint64) ([]uint64, error) {
	return s.tripRepo.GetPlaceIDs(ctx, tripID)
}

// AddPlaceToTrip – добавление места в поездку (требует прав редактора или владельца)
func (s *tripService) AddPlaceToTrip(ctx context.Context, tripID, placeID, userID uint64, orderIndex int16) error {
	logger.Info(ctx, "AddPlaceToTrip called", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	ok, err := s.memberRepo.HasEditPermission(ctx, tripID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not authorized to edit this trip")
	}
	exists, err := s.tripRepo.CheckPlaceInTrip(ctx, tripID, placeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("place already in trip")
	}
	return s.tripRepo.AddAttraction(ctx, tripID, placeID, orderIndex)
}

// RemovePlaceFromTrip – удаление места из поездки (требует прав редактора или владельца)
func (s *tripService) RemovePlaceFromTrip(ctx context.Context, tripID, placeID, userID uint64) error {
	logger.Info(ctx, "RemovePlaceFromTrip called", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	ok, err := s.memberRepo.HasEditPermission(ctx, tripID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not authorized to edit this trip")
	}
	return s.tripRepo.RemoveAttraction(ctx, tripID, placeID)
}

// ----- Методы шеринга и совместного планирования -----

func (s *tripService) CreateViewShareLink(ctx context.Context, tripID, userID uint64) (string, error) {
	role, err := s.memberRepo.GetMemberRole(ctx, tripID, userID)
	if err != nil {
		return "", err
	}
	if role != "owner" {
		return "", errors.New("only owner can create share links")
	}
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	invite := &models.TripInvite{
		TripID:    tripID,
		Token:     token,
		Role:      "viewer",
		IsOneTime: false,
		CreatedBy: userID,
	}
	if err := s.inviteRepo.CreateInvite(ctx, invite); err != nil {
		return "", err
	}
	baseURL := getShareBaseURL()
	return baseURL + "/share/view/" + token, nil
}

func (s *tripService) CreateEditShareLink(ctx context.Context, tripID, userID uint64) (string, error) {
	role, err := s.memberRepo.GetMemberRole(ctx, tripID, userID)
	if err != nil {
		return "", err
	}
	if role != "owner" {
		return "", errors.New("only owner can create share links")
	}
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	invite := &models.TripInvite{
		TripID:    tripID,
		Token:     token,
		Role:      "editor",
		IsOneTime: true,
		CreatedBy: userID,
	}
	if err := s.inviteRepo.CreateInvite(ctx, invite); err != nil {
		return "", err
	}
	baseURL := getShareBaseURL()
	return baseURL + "/share/edit/" + token, nil
}

func (s *tripService) AcceptInvite(ctx context.Context, token string, userID uint64) (tripID uint64, role string, err error) {
	invite, err := s.inviteRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return 0, "", err
	}
	if invite == nil {
		return 0, "", errors.New("invalid or expired invite")
	}
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		return 0, "", errors.New("invite has expired")
	}
	if invite.IsOneTime && invite.UsedAt != nil {
		return 0, "", errors.New("invite already used")
	}
	if err := s.memberRepo.AddMember(ctx, invite.TripID, userID, invite.Role); err != nil {
		return 0, "", err
	}
	if invite.IsOneTime {
		_ = s.inviteRepo.MarkUsed(ctx, invite.ID)
	}
	return invite.TripID, invite.Role, nil
}

func (s *tripService) GetTripMembers(ctx context.Context, tripID, userID uint64) ([]models.TripMember, error) {
	role, err := s.memberRepo.GetMemberRole(ctx, tripID, userID)
	if err != nil {
		return nil, err
	}
	if role != "owner" {
		return nil, errors.New("only owner can view members")
	}
	return s.memberRepo.GetTripMembers(ctx, tripID)
}

func (s *tripService) RemoveMember(ctx context.Context, tripID, ownerID, memberID uint64) error {
	role, err := s.memberRepo.GetMemberRole(ctx, tripID, ownerID)
	if err != nil {
		return err
	}
	if role != "owner" {
		return errors.New("only owner can remove members")
	}
	if ownerID == memberID {
		return errors.New("cannot remove owner")
	}
	return s.memberRepo.RemoveMember(ctx, tripID, memberID)
}

func (s *tripService) GetTripByShareToken(ctx context.Context, token string) (*models.Trip, string, error) {
	invite, err := s.inviteRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, "", err
	}
	if invite == nil {
		return nil, "", errors.New("invalid share token")
	}
	if invite.IsOneTime && invite.UsedAt != nil {
		return nil, "", errors.New("one-time link already used")
	}
	trip, err := s.tripRepo.GetByID(ctx, invite.TripID)
	if err != nil {
		return nil, "", err
	}
	if trip == nil {
		return nil, "", errors.New("trip not found")
	}
	return trip, invite.Role, nil
}
