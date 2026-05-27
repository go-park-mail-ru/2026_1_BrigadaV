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
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
)

type TripHandler struct {
	tripService service.TripService
}

func NewTripHandler(tripService service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	trips, err := h.tripService.GetUserTripsWithRoles(r.Context(), userID)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch trips", logrus.Fields{"error": err})
		writeJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Error: "failed to fetch trips"})
		return
	}
	items := make([]dto.TripResponse, len(trips))
	for i, t := range trips {
		items[i] = dto.TripResponse{
			ID:          t.Trip.ID,
			Title:       t.Trip.Title,
			Location:    t.Trip.Location,
			StartDate:   t.Trip.StartDate,
			EndDate:     t.Trip.EndDate,
			Description: t.Trip.Description,
			Preview:     t.Trip.PreviewURL,
			Role:        t.Role,
		}
	}
	// TripResponseList оборачивает в {"items":[...]}.
	// Тесты ожидают плоский массив — сериализуем items напрямую через json,
	// т.к. []dto.TripResponse не является easyjson.Marshaler.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		logger.Error(r.Context(), "json encode trips error", logrus.Fields{"error": err})
	}
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.CreateTripRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in CreateTrip", logrus.Fields{"error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid request"})
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
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	logger.Info(r.Context(), "Trip created", logrus.Fields{"trip_id": trip.ID})
	writeJSON(w, http.StatusCreated, &dto.CreateTripResponse{ID: trip.ID, Preview: trip.PreviewURL})
}

func (h *TripHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "missing trip id"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in GetDetails", logrus.Fields{"id": idStr, "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	trip, places, role, err := h.tripService.GetTripDetailsWithRole(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "GetTripDetails failed", logrus.Fields{"error": err, "trip_id": id})
		if err.Error() == "trip not found" {
			writeJSON(w, http.StatusNotFound, &dto.ErrorResponse{Error: "trip not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Error: "internal server error"})
		return
	}
	response := dto.TripDetailsResponse{
		ID:          trip.ID,
		Title:       trip.Title,
		Location:    trip.Location,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		Preview:     trip.PreviewURL,
		Attractions: places,
		Role:        role,
	}
	writeJSON(w, http.StatusOK, &response)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in Update", logrus.Fields{"id": vars["id"], "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	var req dto.UpdateTripRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in UpdateTrip", logrus.Fields{"error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid request"})
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
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, &dto.MessageResponse{Message: "ok"})
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
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

func (h *TripHandler) GetTripPlaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in GetTripPlaces", logrus.Fields{"id": vars["id"], "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	placeIDs, err := h.tripService.GetTripPlaceIDs(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "Failed to fetch place IDs", logrus.Fields{"error": err, "trip_id": id})
		writeJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Error: "failed to fetch place IDs"})
		return
	}
	if placeIDs == nil {
		placeIDs = []uint64{}
	}
	// TripPlacesResponse — easyjson тип-алиас для []uint64
	resp := dto.TripPlacesResponse(placeIDs)
	w.Header().Set("Content-Type", "application/json")
	data, err := easyjson.Marshal(resp)
	if err != nil {
		logger.Error(r.Context(), "easyjson marshal error", logrus.Fields{"error": err})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (h *TripHandler) AddPlace(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in AddPlace", logrus.Fields{"id": vars["id"], "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	var req dto.AddPlaceRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		logger.Error(r.Context(), "Invalid JSON in AddPlace", logrus.Fields{"error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid request"})
		return
	}
	if err := h.tripService.AddPlaceToTrip(r.Context(), tripID, req.PlaceID, userID, req.OrderIndex); err != nil {
		logger.Error(r.Context(), "AddPlaceToTrip failed", logrus.Fields{"error": err, "trip_id": tripID, "place_id": req.PlaceID})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	logger.Info(r.Context(), "Place added to trip", logrus.Fields{"trip_id": tripID, "place_id": req.PlaceID})
	writeJSON(w, http.StatusOK, &dto.MessageResponse{Message: "place added to trip"})
}

func (h *TripHandler) RemovePlace(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in RemovePlace", logrus.Fields{"id": vars["id"], "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	placeID, err := strconv.ParseUint(vars["placeId"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid place id in RemovePlace", logrus.Fields{"placeId": vars["placeId"], "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid place id"})
		return
	}
	if err := h.tripService.RemovePlaceFromTrip(r.Context(), tripID, placeID, userID); err != nil {
		logger.Error(r.Context(), "RemovePlaceFromTrip failed", logrus.Fields{"error": err, "trip_id": tripID, "place_id": placeID})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	logger.Info(r.Context(), "Place removed from trip", logrus.Fields{"trip_id": tripID, "place_id": placeID})
	w.WriteHeader(http.StatusNoContent)
}

func (h *TripHandler) CreateViewShareLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	link, err := h.tripService.CreateViewShareLink(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "CreateViewShareLink failed", logrus.Fields{"error": err})
		writeJSON(w, http.StatusForbidden, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, &dto.ShareLinkResponse{ShareLink: link})
}

func (h *TripHandler) CreateEditShareLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	link, err := h.tripService.CreateEditShareLink(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "CreateEditShareLink failed", logrus.Fields{"error": err})
		writeJSON(w, http.StatusForbidden, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, &dto.ShareLinkResponse{ShareLink: link})
}

func (h *TripHandler) AcceptInviteRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
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

func (h *TripHandler) GetTripMembers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	members, err := h.tripService.GetTripMembers(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "GetTripMembers failed", logrus.Fields{"error": err})
		writeJSON(w, http.StatusForbidden, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	// Тесты декодируют в []models.TripMember — используем json чтобы сохранить совместимость.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(members); err != nil {
		logger.Error(r.Context(), "json encode error", logrus.Fields{"error": err})
	}
}

func (h *TripHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid trip id in RemoveMember", logrus.Fields{"id": vars["id"], "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	memberIDStr, ok := vars["member_id"]
	if !ok || memberIDStr == "" {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "missing member id"})
		return
	}
	memberID, err := strconv.ParseUint(memberIDStr, 10, 64)
	if err != nil {
		logger.Error(r.Context(), "Invalid member id in RemoveMember", logrus.Fields{"member_id": memberIDStr, "error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid member id"})
		return
	}
	err = h.tripService.RemoveMember(r.Context(), tripID, userID, memberID)
	if err != nil {
		logger.Error(r.Context(), "RemoveMember failed", logrus.Fields{"error": err})
		writeJSON(w, http.StatusForbidden, &dto.ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
	response := dto.SharedTripResponse{
		Trip:        trip,
		Attractions: places,
		Role:        role,
	}
	writeJSON(w, http.StatusOK, &response)
}

func (h *TripHandler) ExportTripToPDF(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid trip id"})
		return
	}
	pdfData, err := h.tripService.ExportTripToPDF(r.Context(), id, userID)
	if err != nil {
		logger.Error(r.Context(), "ExportTripToPDF failed", logrus.Fields{"error": err, "trip_id": id})
		switch err.Error() {
		case "access denied":
			http.Error(w, "access denied", http.StatusForbidden)
		case "trip not found":
			http.Error(w, "trip not found", http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=trip_%d.pdf", id))
	w.Write(pdfData)
}

// AcceptInviteAPI – POST /api/share/edit/{token} – JSON-ответ для фронтенда.
func (h *TripHandler) AcceptInviteAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	tripID, _, err := h.tripService.AcceptInvite(r.Context(), token, userID)
	if err != nil {
		logger.Error(r.Context(), "AcceptInviteAPI failed", logrus.Fields{"error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid share link"})
		return
	}
	writeJSON(w, http.StatusOK, &dto.TripIDResponse{TripID: tripID})
}

// JoinViewShareAPI – POST /api/share/view/{token} – JSON-ответ для фронтенда.
func (h *TripHandler) JoinViewShareAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	tripID, _, err := h.tripService.AcceptInvite(r.Context(), token, userID)
	if err != nil {
		logger.Error(r.Context(), "JoinViewShareAPI failed", logrus.Fields{"error": err})
		writeJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Error: "invalid share link"})
		return
	}
	writeJSON(w, http.StatusOK, &dto.TripIDResponse{TripID: tripID})
}
