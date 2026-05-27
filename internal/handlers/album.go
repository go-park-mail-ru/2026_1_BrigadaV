package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	pb "guidely-app/pkg/pb/album"
	"guidely-app/pkg/storage"

	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
)

const (
	uploadDir    = "./uploads/photos"
	maxPhotoSize = 30 << 20 // 30 MB
)

type AlbumHandler struct {
	client pb.AlbumServiceClient
	s3     *storage.S3Client
}

func NewAlbumHandler(client pb.AlbumServiceClient, s3 *storage.S3Client) *AlbumHandler {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		logrus.Warnf("failed to create upload dir: %v", err)
	}
	return &AlbumHandler{client: client, s3: s3}
}

func writeJSON(w http.ResponseWriter, status int, v easyjson.Marshaler) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := easyjson.MarshalToWriter(v, w); err != nil {
		logrus.Errorf("easyjson marshal error: %v", err)
	}
}

func (h *AlbumHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.CreateAlbumRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Create(r.Context(), &pb.CreateAlbumRequest{
		TripId:      req.TripID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		logger.Error(r.Context(), "album create error", logrus.Fields{"error": err, "trip_id": req.TripID})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	albumResp := dto.AlbumResponse{
		ID:          resp.Id,
		TripID:      resp.TripId,
		Name:        resp.Name,
		Description: resp.Description,
	}
	writeJSON(w, http.StatusCreated, &albumResp)
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
		logger.Error(r.Context(), "album get error", logrus.Fields{"error": err, "album_id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	albumResp := dto.AlbumResponse{
		ID:          resp.Id,
		TripID:      resp.TripId,
		Name:        resp.Name,
		Description: resp.Description,
	}
	writeJSON(w, http.StatusOK, &albumResp)
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
		logger.Warn(r.Context(), "album not found for trip, auto-creating", logrus.Fields{
			"trip_id": tripID,
			"error":   err,
		})
		created, cerr := h.client.Create(r.Context(), &pb.CreateAlbumRequest{
			TripId:    tripID,
			Name:      "Основной альбом",
			MaxPhotos: 50,
		})
		if cerr != nil {
			logger.Error(r.Context(), "album auto-create error", logrus.Fields{"error": cerr, "trip_id": tripID})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		albumResp := dto.AlbumResponse{
			ID:     created.Id,
			TripID: created.TripId,
			Name:   created.Name,
		}
		writeJSON(w, http.StatusOK, &albumResp)
		return
	}

	albumResp := dto.AlbumResponse{
		ID:          resp.Id,
		TripID:      resp.TripId,
		Name:        resp.Name,
		Description: resp.Description,
	}
	writeJSON(w, http.StatusOK, &albumResp)
}

func (h *AlbumHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	var req dto.UpdateAlbumRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Update(r.Context(), &pb.UpdateAlbumRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		logger.Error(r.Context(), "album update error", logrus.Fields{"error": err, "album_id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	albumResp := dto.AlbumResponse{
		ID:          resp.Id,
		TripID:      resp.TripId,
		Name:        resp.Name,
		Description: resp.Description,
	}
	writeJSON(w, http.StatusOK, &albumResp)
}

func (h *AlbumHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	if _, err := h.client.Delete(r.Context(), &pb.DeleteAlbumRequest{Id: id}); err != nil {
		logger.Error(r.Context(), "album delete error", logrus.Fields{"error": err, "album_id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddPhoto — POST /api/albums/{id}/photos
func (h *AlbumHandler) AddPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoSize+1<<20)
	if err := r.ParseMultipartForm(maxPhotoSize); err != nil {
		if strings.Contains(err.Error(), "too large") || strings.Contains(err.Error(), "request body too large") {
			http.Error(w, "file too large (max 30 MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo field missing", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxPhotoSize {
		http.Error(w, "file too large (max 30 MB)", http.StatusRequestEntityTooLarge)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d_%d%s", albumID, time.Now().UnixNano(), ext)

	var fileURL string

	if h.s3 != nil {
		objectName := "photos/" + filename
		fileURL, err = h.s3.UploadFile(r.Context(), objectName, file, header.Size, contentType)
		if err != nil {
			logger.Error(r.Context(), "S3 upload failed", logrus.Fields{"error": err, "album_id": albumID})
			http.Error(w, "internal error: failed to upload photo", http.StatusInternalServerError)
			return
		}
		logger.Info(r.Context(), "photo uploaded to S3", logrus.Fields{"url": fileURL, "album_id": albumID})
	} else {
		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			logger.Error(r.Context(), "mkdir error", logrus.Fields{"error": err, "dir": uploadDir})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		savePath := filepath.Join(uploadDir, filename)
		dst, err := os.Create(savePath)
		if err != nil {
			logger.Error(r.Context(), "file create error", logrus.Fields{"error": err, "path": savePath})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			logger.Error(r.Context(), "file copy error", logrus.Fields{"error": err})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		fileURL = "/uploads/photos/" + filename
		logger.Info(r.Context(), "photo saved locally", logrus.Fields{"path": fileURL, "album_id": albumID})
	}

	addResp, err := h.client.UploadPhoto(r.Context(), &pb.UploadPhotoRequest{
		AlbumId:  albumID,
		FilePath: fileURL,
	})
	if err != nil {
		logger.Error(r.Context(), "album upload photo gRPC error", logrus.Fields{
			"error":    err,
			"album_id": albumID,
		})
		if h.s3 == nil {
			os.Remove(filepath.Join(uploadDir, filename))
		}
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(r.Context(), "photo added to album", logrus.Fields{
		"album_id": albumID,
		"photo_id": addResp.PhotoId,
		"url":      fileURL,
	})

	resp := dto.PhotoUploadResponse{
		ID:  addResp.PhotoId,
		URL: fileURL,
	}
	writeJSON(w, http.StatusCreated, &resp)
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
		logger.Error(r.Context(), "album remove photo error", logrus.Fields{
			"error":    err,
			"album_id": albumID,
			"photo_id": photoID,
		})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetPhotos — GET /api/albums/{id}/photos
func (h *AlbumHandler) GetPhotos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetPhotos(r.Context(), &pb.GetAlbumPhotosRequest{AlbumId: albumID})
	if err != nil {
		logger.Error(r.Context(), "album get photos error", logrus.Fields{"error": err, "album_id": albumID})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	result := make(dto.PhotoUploadResponseList, 0, len(resp.Photos))
	for _, p := range resp.Photos {
		result = append(result, dto.PhotoUploadResponse{
			ID:  p.PhotoId,
			URL: p.FileUrl,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := easyjson.Marshal(result)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
