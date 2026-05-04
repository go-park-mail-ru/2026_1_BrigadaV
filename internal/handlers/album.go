package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"guidely-app/internal/middleware"
	pb "guidely-app/pkg/pb/album"

	"github.com/gorilla/mux"
)

const uploadDir = "./uploads/photos"

type AlbumHandler struct {
	client pb.AlbumServiceClient
}

func NewAlbumHandler(client pb.AlbumServiceClient) *AlbumHandler {
	return &AlbumHandler{client: client}
}

// photoResponse — то что ждёт фронтенд: { id, url }
type photoResponse struct {
	ID  uint64 `json:"id"`
	URL string `json:"url"`
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

	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetByTrip — GET /api/trips/{tripID}/album
// Возвращает альбом. Триггер в БД создаёт альбом автоматически при создании поездки.
// Если по каким-то причинам альбома нет — создаём на лету.
func (h *AlbumHandler) GetByTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["tripID"], 10, 64)
	if err != nil {
		http.Error(w, "invalid trip id", http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetByTrip(r.Context(), &pb.GetAlbumByTripRequest{TripId: tripID})
	if err != nil {
		// Альбома нет — создаём автоматически
		log.Printf("album not found for trip %d, auto-creating: %v", tripID, err)
		created, cerr := h.client.Create(r.Context(), &pb.CreateAlbumRequest{
			TripId:    tripID,
			Name:      "Основной альбом",
			MaxPhotos: 50,
		})
		if cerr != nil {
			log.Printf("album auto-create error: %v", cerr)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(created)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
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

// AddPhoto — POST /api/albums/{id}/photos
// Принимает multipart/form-data с полем "photo" (файл).
// Сохраняет файл на диск и регистрирует через gRPC UploadPhoto.
func (h *AlbumHandler) AddPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	// Ограничение — 10 МБ
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo field missing", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Создаём папку если нет
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Printf("mkdir error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Уникальное имя файла
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d_%d%s", albumID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)
	relativePath := "uploads/photos/" + filename

	// Сохраняем файл
	dst, err := os.Create(savePath)
	if err != nil {
		log.Printf("file create error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("file copy error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Регистрируем фото и связываем с альбомом через gRPC
	addResp, err := h.client.UploadPhoto(r.Context(), &pb.UploadPhotoRequest{
		AlbumId:  albumID,
		FilePath: relativePath,
	})
	if err != nil {
		log.Printf("album upload photo gRPC error: %v", err)
		os.Remove(savePath)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(photoResponse{
		ID:  addResp.PhotoId,
		URL: relativePath,
	})
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

// GetPhotos — GET /api/albums/{id}/photos
// Возвращает массив { id, url } — именно это ожидает фронтенд.
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

	result := make([]photoResponse, 0, len(resp.Photos))
	for _, p := range resp.Photos {
		result = append(result, photoResponse{
			ID:  p.PhotoId,
			URL: p.FileUrl,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
