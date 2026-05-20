package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	"guidely-app/internal/service"
	"guidely-app/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type TripHandler struct {
	tripService service.TripService
}

func NewTripHandler(tripService service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

// List возвращает поездки, где пользователь является создателем (или участником – при расширении)
func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	trips, err := h.tripService.GetUserTrips(r.Context(), userID)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch trips", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch trips"})
		return
	}
	response := make([]dto.TripResponse, len(trips))
	for i, t := range trips {
		response[i] = dto.TripResponse{
			ID:          t.ID,
			Title:       t.Title,
			Location:    t.Location,
			StartDate:   t.StartDate,
			EndDate:     t.EndDate,
			Description: t.Description,
			Preview:     t.PreviewURL,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Create – создание новой поездки
func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	var req dto.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in CreateTrip", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	input := service.CreateTripInput{
		Title:      req.Title,
		Location:   req.Location,
		StartDate:  utils.ParseDatePtr(req.StartDate),
		EndDate:    utils.ParseDatePtr(req.EndDate),
		PreviewURL: req.Preview,
		CreatedBy:  userID,
		IsPublic:   req.IsPublic,
	}
	trip, err := h.tripService.Create(r.Context(), input)
	if err != nil {
		logger.Error(r.Context(), "CreateTrip failed", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	logger.Info(r.Context(), "Trip created", logrus.Fields{"trip_id": trip.ID})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateTripResponse{
		ID:      trip.ID,
		Preview: trip.PreviewURL,
	})
}

// GetDetails – детали поездки (требует права просмотра)
func (h *TripHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing trip id"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in GetDetails", logrus.Fields{"id": idStr, "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	trip, places, err := h.tripService.GetTripDetails(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "GetTripDetails failed", logrus.Fields{"error": err, "trip_id": id})
		if err.Error() == "trip not found" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "trip not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	// Проверка прав просмотра – в сервисе она уже сделана (через memberRepo)
	response := dto.TripDetailsResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Location:    trip.Location,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		Preview:     trip.PreviewURL,
		Attractions: places,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Update – обновление поездки (требует прав редактора)
func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in Update", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	var req dto.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in UpdateTrip", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	input := service.UpdateTripInput{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartDate:   utils.ParseDatePtr(req.StartDate),
		EndDate:     utils.ParseDatePtr(req.EndDate),
		PreviewURL:  req.Preview,
		IsPublic:    req.IsPublic,
	}
	_, err = h.tripService.Update(r.Context(), id, userID, input)
	if err != nil {
		logger.Error(r.Context(), "UpdateTrip failed", logrus.Fields{"error": err, "trip_id": id})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
}

// Delete – удаление поездки (только владелец)
func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	if err := h.tripService.Delete(r.Context(), id, userID); err != nil {
		logger.Error(r.Context(), "trip delete error", logrus.Fields{"error": err, "trip_id": id})
		switch {
		case err.Error() == "trip not found":
			http.Error(w, "trip not found", http.StatusNotFound)
		case err.Error() == "only owner can delete trip":
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetTripPlaces – список ID достопримечательностей в поездке
func (h *TripHandler) GetTripPlaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in GetTripPlaces", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	placeIDs, err := h.tripService.GetTripPlaceIDs(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch place IDs", logrus.Fields{"error": err, "trip_id": id})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch place IDs"})
		return
	}
	if placeIDs == nil {
		placeIDs = []uint64{}
	}
	json.NewEncoder(w).Encode(placeIDs)
}

// AddPlace – добавление места в поездку (требует прав редактора)
func (h *TripHandler) AddPlace(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in AddPlace", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	var req struct {
		PlaceID    uint64 `json:"place_id"`
		OrderIndex int16  `json:"order_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in AddPlace", logrus.Fields{"error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	if err := h.tripService.AddPlaceToTrip(r.Context(), tripID, req.PlaceID, userID, req.OrderIndex); err != nil {
		logger.Error(r.Context(), "AddPlaceToTrip failed", logrus.Fields{"error": err, "trip_id": tripID, "place_id": req.PlaceID})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	logger.Info(r.Context(), "Place added to trip", logrus.Fields{"trip_id": tripID, "place_id": req.PlaceID})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "place added to trip"})
}

// RemovePlace – удаление места из поездки (требует прав редактора)
func (h *TripHandler) RemovePlace(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in RemovePlace", logrus.Fields{"id": vars["id"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid trip id"})
		return
	}
	placeID, err := strconv.ParseUint(vars["placeId"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid place id in RemovePlace", logrus.Fields{"placeId": vars["placeId"], "error": err})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid place id"})
		return
	}
	if err := h.tripService.RemovePlaceFromTrip(r.Context(), tripID, placeID, userID); err != nil {
		logger.Error(r.Context(), "RemovePlaceFromTrip failed", logrus.Fields{"error": err, "trip_id": tripID, "place_id": placeID})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	logger.Info(r.Context(), "Place removed from trip", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	w.WriteHeader(http.StatusNoContent)
}

// ----- Методы шеринга -----

// CreateViewShareLink – постоянная ссылка для просмотра
func (h *TripHandler) CreateViewShareLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid trip id", http.StatusBadRequest)
		return
	}
	link, err := h.tripService.CreateViewShareLink(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "CreateViewShareLink failed", logrus.Fields{"error": err})
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"share_link": link})
}

// CreateEditShareLink – одноразовая ссылка для редактирования
func (h *TripHandler) CreateEditShareLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid trip id", http.StatusBadRequest)
		return
	}
	link, err := h.tripService.CreateEditShareLink(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "CreateEditShareLink failed", logrus.Fields{"error": err})
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"share_link": link})
}

// AcceptInviteRedirect – GET /api/share/edit/{token} – принимает приглашение и редиректит на страницу поездки
func (h *TripHandler) AcceptInviteRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		// Сохраняем токен в сессию или параметр редиректа
		http.Redirect(w, r, "/login?redirect=/share/edit/"+token, http.StatusFound)
		return
	}
	tripID, role, err := h.tripService.AcceptInvite(r.Context(), token, userID)
	if err != nil {
		logger.Error(r.Context(), "AcceptInvite failed", logrus.Fields{"error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	redirectURL := fmt.Sprintf("/trips/%d?role=%s", tripID, role)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// GetTripMembers – GET /api/trips/{id}/members – список участников (только для владельца)
func (h *TripHandler) GetTripMembers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid trip id", http.StatusBadRequest)
		return
	}
	members, err := h.tripService.GetTripMembers(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "GetTripMembers failed", logrus.Fields{"error": err})
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// RemoveMember – DELETE /api/trips/{id}/members/{member_id} – удаление участника (только владелец)
func (h *TripHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid trip id", http.StatusBadRequest)
		return
	}
	memberID, err := strconv.ParseUint(vars["member_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid member id", http.StatusBadRequest)
		return
	}
	err = h.tripService.RemoveMember(r.Context(), tripID, userID, memberID)
	if err != nil {
		logger.Error(r.Context(), "RemoveMember failed", logrus.Fields{"error": err})
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ViewSharedTrip – GET /api/share/view/{token} – публичный просмотр поездки по ссылке
func (h *TripHandler) ViewSharedTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	trip, role, err := h.tripService.GetTripByShareToken(r.Context(), token)
	if err != nil {
		logger.Error(r.Context(), "ViewSharedTrip failed", logrus.Fields{"error": err})
		http.Error(w, "invalid share link", http.StatusNotFound)
		return
	}
	_, places, err := h.tripService.GetTripDetails(r.Context(), trip.ID)
	if err != nil {
		logger.Error(r.Context(), "Failed to get attractions for shared trip", logrus.Fields{"error": err})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"trip":        trip,
		"attractions": places,
		"role":        role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
