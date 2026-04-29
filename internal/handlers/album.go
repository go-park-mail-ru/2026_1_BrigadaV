package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"guidely-app/internal/models"
	"guidely-app/internal/service"

	"github.com/gorilla/mux"
)

type AlbumHandler struct {
	svc service.AlbumService
}

func NewAlbumHandler(svc service.AlbumService) *AlbumHandler {
	return &AlbumHandler{svc: svc}
}

func (h *AlbumHandler) Create(w http.ResponseWriter, r *http.Request) {
	var album models.Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.svc.Create(r.Context(), &album); err != nil {
		log.Printf("album create error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(album)
}

func (h *AlbumHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}
	album, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("album get error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if album == nil {
		http.Error(w, "album not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(album)
}

func (h *AlbumHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}
	var album models.Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	album.ID = id
	if err := h.svc.Update(r.Context(), &album); err != nil {
		log.Printf("album update error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(album)
}

func (h *AlbumHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		log.Printf("album delete error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
