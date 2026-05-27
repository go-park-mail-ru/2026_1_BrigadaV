package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"guidely-app/internal/logger"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/sirupsen/logrus"
)

type UserTripInfo struct {
	Trip models.Trip
	Role string
}

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
	// Приоритет: SHARE_BASE_URL → FRONTEND_URL → fallback
	if base := os.Getenv("SHARE_BASE_URL"); base != "" {
		return base
	}
	if front := os.Getenv("FRONTEND_URL"); front != "" {
		return front
	}
	return "http://localhost:8080"
}

// isOwner проверяет что пользователь является владельцем поездки.
// Сначала смотрим в trip_member, если там нет записи — проверяем created_by.
// Это нужно для обратной совместимости с поездками, созданными до добавления
// автоматической записи owner в trip_member.
func (s *tripService) isOwner(ctx context.Context, tripID, userID uint64) (bool, error) {
	role, err := s.memberRepo.GetMemberRole(ctx, tripID, userID)
	if err != nil {
		return false, err
	}
	if role == "owner" {
		return true, nil
	}
	// Fallback: проверяем created_by напрямую
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return false, err
	}
	if trip == nil {
		return false, errors.New("trip not found")
	}
	return trip.CreatedBy == userID, nil
}

// Create – создание новой поездки
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
	logger.Info(ctx, "CreateTrip successful", logrus.Fields{"trip_id": trip.ID})
	return trip, nil
}

// GetUserTrips – поездки, где пользователь является участником
func (s *tripService) GetUserTrips(ctx context.Context, userID uint64) ([]models.Trip, error) {
	return s.tripRepo.GetByUser(ctx, userID)
}

// GetTripDetails – получение поездки и её достопримечательностей
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

// Update – обновление поездки (владелец или companion)
func (s *tripService) Update(ctx context.Context, id, userID uint64, input UpdateTripInput) (*models.Trip, error) {
	logger.Info(ctx, "UpdateTrip called", logrus.Fields{"trip_id": id, "user_id": userID})

	ok, err := s.memberRepo.HasEditPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		// Fallback: проверяем created_by
		trip, err2 := s.tripRepo.GetByID(ctx, id)
		if err2 != nil || trip == nil || trip.CreatedBy != userID {
			return nil, errors.New("not authorized to edit this trip")
		}
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

	owner, err := s.isOwner(ctx, id, userID)
	if err != nil {
		return err
	}
	if !owner {
		return errors.New("only owner can delete trip")
	}
	return s.tripRepo.Delete(ctx, id)
}

// GetTripPlaceIDs – список ID достопримечательностей
func (s *tripService) GetTripPlaceIDs(ctx context.Context, tripID uint64) ([]uint64, error) {
	return s.tripRepo.GetPlaceIDs(ctx, tripID)
}

// AddPlaceToTrip – добавление места (владелец или companion)
func (s *tripService) AddPlaceToTrip(ctx context.Context, tripID, placeID, userID uint64, orderIndex int16) error {
	logger.Info(ctx, "AddPlaceToTrip called", logrus.Fields{"trip_id": tripID, "place_id": placeID})

	ok, err := s.memberRepo.HasEditPermission(ctx, tripID, userID)
	if err != nil {
		return err
	}
	if !ok {
		// Fallback: проверяем created_by
		trip, err2 := s.tripRepo.GetByID(ctx, tripID)
		if err2 != nil || trip == nil || trip.CreatedBy != userID {
			return errors.New("not authorized to edit this trip")
		}
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

// RemovePlaceFromTrip – удаление места (владелец или companion)
func (s *tripService) RemovePlaceFromTrip(ctx context.Context, tripID, placeID, userID uint64) error {
	logger.Info(ctx, "RemovePlaceFromTrip called", logrus.Fields{"trip_id": tripID, "place_id": placeID})

	ok, err := s.memberRepo.HasEditPermission(ctx, tripID, userID)
	if err != nil {
		return err
	}
	if !ok {
		trip, err2 := s.tripRepo.GetByID(ctx, tripID)
		if err2 != nil || trip == nil || trip.CreatedBy != userID {
			return errors.New("not authorized to edit this trip")
		}
	}
	return s.tripRepo.RemoveAttraction(ctx, tripID, placeID)
}

// ----- Методы шеринга и совместного планирования -----

// CreateViewShareLink создаёт постоянную ссылку для просмотра поездки.
// Доступно только владельцу.
func (s *tripService) CreateViewShareLink(ctx context.Context, tripID, userID uint64) (string, error) {
	owner, err := s.isOwner(ctx, tripID, userID)
	if err != nil {
		return "", err
	}
	if !owner {
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

// CreateEditShareLink создаёт одноразовую ссылку для добавления редактора.
// Доступно только владельцу.
func (s *tripService) CreateEditShareLink(ctx context.Context, tripID, userID uint64) (string, error) {
	owner, err := s.isOwner(ctx, tripID, userID)
	if err != nil {
		return "", err
	}
	if !owner {
		return "", errors.New("only owner can create share links")
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}
	invite := &models.TripInvite{
		TripID:    tripID,
		Token:     token,
		Role:      "companion",
		IsOneTime: true,
		CreatedBy: userID,
	}
	if err := s.inviteRepo.CreateInvite(ctx, invite); err != nil {
		return "", err
	}
	baseURL := getShareBaseURL()
	return baseURL + "/share/edit/" + token, nil
}

// AcceptInvite принимает приглашение по токену и добавляет пользователя в поездку.
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

// GetTripMembers возвращает список участников (только владелец).
func (s *tripService) GetTripMembers(ctx context.Context, tripID, userID uint64) ([]models.TripMember, error) {
	owner, err := s.isOwner(ctx, tripID, userID)
	if err != nil {
		return nil, err
	}
	if !owner {
		return nil, errors.New("only owner can view members")
	}
	return s.memberRepo.GetTripMembers(ctx, tripID)
}

// RemoveMember удаляет участника из поездки (только владелец).
func (s *tripService) RemoveMember(ctx context.Context, tripID, ownerID, memberID uint64) error {
	owner, err := s.isOwner(ctx, tripID, ownerID)
	if err != nil {
		return err
	}
	if !owner {
		return errors.New("only owner can remove members")
	}
	if ownerID == memberID {
		return errors.New("cannot remove owner")
	}
	return s.memberRepo.RemoveMember(ctx, tripID, memberID)
}

// GetTripByShareToken возвращает поездку по токену приглашения (публичный просмотр).
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

func formatDatePtr(t *time.Time) string {
	if t == nil {
		return "не указана"
	}
	return t.Format("02.01.2006")
}

func (s *tripService) ExportTripToPDF(ctx context.Context, tripID, userID uint64) ([]byte, error) {
	ok, err := s.memberRepo.HasViewPermission(ctx, tripID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("access denied")
	}

	trip, places, err := s.GetTripDetails(ctx, tripID)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddUTF8Font("PTSans", "", "assets/fonts/PTSans-Regular.ttf")

	pdf.AddPage()

	pdf.SetFont("PTSans", "", 16)

	title := "Поездка: " + trip.Title
	pdf.Cell(0, 10, title)
	pdf.Ln(12)

	pdf.SetFont("PTSans", "", 12)
	pdf.Cell(40, 10, "Направление:")

	location := "не указано"
	if trip.Location != nil && *trip.Location != "" {
		location = *trip.Location
	}
	pdf.Cell(0, 10, location)
	pdf.Ln(8)

	pdf.Cell(40, 10, "Даты:")
	dateFrom := formatDatePtr(trip.StartDate)
	dateTo := formatDatePtr(trip.EndDate)
	dateStr := dateFrom
	if dateFrom != "не указана" && dateTo != "не указана" && dateFrom != dateTo {
		dateStr += " – " + dateTo
	} else if dateTo != "не указана" {
		dateStr = dateTo
	}
	pdf.Cell(0, 10, dateStr)
	pdf.Ln(8)

	if trip.Description != "" {
		pdf.Cell(40, 10, "Описание:")
		pdf.Ln(6)
		pdf.MultiCell(0, 6, trip.Description, "", "", false)
		pdf.Ln(4)
	}

	pdf.SetFont("PTSans", "", 14)
	pdf.Cell(0, 10, "Достопримечательности:")
	pdf.Ln(10)

	if len(places) == 0 {
		pdf.SetFont("PTSans", "", 12)
		pdf.Cell(0, 10, "Нет добавленных мест")
		pdf.Ln(10)
	} else {
		for i, place := range places {
			// Проверяем, нужна ли новая страница
			if pdf.GetY() > 250 {
				pdf.AddPage()
			}

			// Название достопримечательности
			pdf.SetFont("PTSans", "", 12)
			pdf.Cell(0, 8, fmt.Sprintf("%d. %s", i+1, place.Name))
			pdf.Ln(7)

			// Рейтинг
			if place.Rating > 0 {
				pdf.SetFont("PTSans", "", 10)
				pdf.Cell(10, 5, "")
				pdf.Cell(0, 5, fmt.Sprintf("Рейтинг: %.1f / 5.0", place.Rating))
				pdf.Ln(6)
			}

			// Описание
			if place.Description != "" {
				pdf.SetFont("PTSans", "", 10)
				desc := place.Description
				if len([]rune(desc)) > 300 {
					runes := []rune(desc)
					desc = string(runes[:300]) + "..."
				}
				pdf.Cell(10, 5, "")
				pdf.MultiCell(0, 5, desc, "", "", false)
				pdf.Ln(1)
			}

			pdf.Ln(3)
		}
	}

	var pdfBuffer bytes.Buffer
	if err := pdf.Output(&pdfBuffer); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return pdfBuffer.Bytes(), nil
}

// GetUserTripsWithRoles возвращает все поездки пользователя с его ролью.
func (s *tripService) GetUserTripsWithRoles(ctx context.Context, userID uint64) ([]UserTripInfo, error) {
	tripsWithRoles, err := s.tripRepo.GetUserTripsWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]UserTripInfo, len(tripsWithRoles))
	for i, tr := range tripsWithRoles {
		result[i] = UserTripInfo{Trip: tr.Trip, Role: tr.Role}
	}
	return result, nil
}

// GetTripDetailsWithRole возвращает поездку, достопримечательности и роль пользователя.
func (s *tripService) GetTripDetailsWithRole(ctx context.Context, tripID, userID uint64) (*models.Trip, []models.PlaceInTrip, string, error) {
	trip, places, err := s.GetTripDetails(ctx, tripID)
	if err != nil {
		return nil, nil, "", err
	}
	role, err := s.tripRepo.GetUserRoleForTrip(ctx, tripID, userID)
	if err != nil {
		return nil, nil, "", err
	}
	return trip, places, role, nil
}
