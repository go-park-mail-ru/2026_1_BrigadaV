package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"guidely-app/internal/middleware"
	pb "guidely-app/pkg/pb/album"

	"github.com/gorilla/mux"
)

type AlbumHandler struct {
	client pb.AlbumServiceClient
}

func NewAlbumHandler(client pb.AlbumServiceClient) *AlbumHandler {
	return &AlbumHandler{client: client}
}

func (h *AlbumHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		TripID      uint64 `json:"trip_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		MaxPhotos   int32  `json:"max_photos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Create(r.Context(), &pb.CreateAlbumRequest{
		TripId:      req.TripID,
		Name:        req.Name,
		Description: req.Description,
		MaxPhotos:   req.MaxPhotos,
	})
	if err != nil {
		log.Printf("album create error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *AlbumHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Get(r.Context(), &pb.GetAlbumRequest{Id: id})
	if err != nil {
		log.Printf("album get error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *AlbumHandler) GetByTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["tripID"], 10, 64)
	if err != nil {
		http.Error(w, "invalid trip id", http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetByTrip(r.Context(), &pb.GetAlbumByTripRequest{TripId: tripID})
	if err != nil {
		log.Printf("album get by trip error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *AlbumHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		MaxPhotos   int32  `json:"max_photos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Update(r.Context(), &pb.UpdateAlbumRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		MaxPhotos:   req.MaxPhotos,
	})
	if err != nil {
		log.Printf("album update error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *AlbumHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	if _, err := h.client.Delete(r.Context(), &pb.DeleteAlbumRequest{Id: id}); err != nil {
		log.Printf("album delete error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AlbumHandler) AddPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	var req struct {
		PhotoID uint64 `json:"photo_id"`
		Order   int32  `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.client.AddPhoto(r.Context(), &pb.AddPhotoRequest{
		AlbumId: albumID,
		PhotoId: req.PhotoID,
		Order:   req.Order,
	})
	if err != nil {
		log.Printf("album add photo error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *AlbumHandler) RemovePhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}
	photoID, err := strconv.ParseUint(vars["photoId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid photo id", http.StatusBadRequest)
		return
	}

	if _, err := h.client.RemovePhoto(r.Context(), &pb.RemovePhotoRequest{
		AlbumId: albumID,
		PhotoId: photoID,
	}); err != nil {
		log.Printf("album remove photo error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AlbumHandler) GetPhotos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetPhotos(r.Context(), &pb.GetAlbumPhotosRequest{AlbumId: albumID})
	if err != nil {
		log.Printf("album get photos error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp.Photos)
}
