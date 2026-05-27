package handlers

import (
	"net/http"
	"strconv"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"guidely-app/pkg/models"

	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
)

type CategoryHandler struct {
	svc service.CategoryService
}

func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.GetAll(r.Context())
	if err != nil {
		logger.Error(r.Context(), "category list error", logrus.Fields{"error": err})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
		return
	}
	response := make(dto.CategoryResponseList, len(categories))
	for i, cat := range categories {
		response[i] = dto.CategoryResponse{
			ID:              cat.ID,
			Name:            cat.Name,
			Description:     cat.Description,
			ApplicableTypes: cat.ApplicableTypes,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	data, err := easyjson.Marshal(response)
	if err != nil {
		logger.Error(r.Context(), "easyjson marshal error", logrus.Fields{"error": err})
		return
	}
	w.Write(data)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}
	cat, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "category get error", logrus.Fields{"error": err, "id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if cat == nil {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	response := dto.CategoryResponse{
		ID:              cat.ID,
		Name:            cat.Name,
		Description:     cat.Description,
		ApplicableTypes: cat.ApplicableTypes,
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := easyjson.MarshalToWriter(&response, w); err != nil {
		logger.Error(r.Context(), "easyjson marshal error", logrus.Fields{"error": err})
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CategoryRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	cat := &models.Category{
		Name:            req.Name,
		Description:     req.Description,
		ApplicableTypes: req.ApplicableTypes,
	}
	if err := h.svc.Create(r.Context(), cat); err != nil {
		logger.Error(r.Context(), "category create error", logrus.Fields{"error": err})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	response := dto.CategoryResponse{
		ID:              cat.ID,
		Name:            cat.Name,
		Description:     cat.Description,
		ApplicableTypes: cat.ApplicableTypes,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := easyjson.MarshalToWriter(&response, w); err != nil {
		logger.Error(r.Context(), "easyjson marshal error", logrus.Fields{"error": err})
	}
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}
	var req dto.CategoryRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	cat := &models.Category{
		ID:              id,
		Name:            req.Name,
		Description:     req.Description,
		ApplicableTypes: req.ApplicableTypes,
	}
	if err := h.svc.Update(r.Context(), cat); err != nil {
		logger.Error(r.Context(), "category update error", logrus.Fields{"error": err, "id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	response := dto.CategoryResponse{
		ID:              cat.ID,
		Name:            cat.Name,
		Description:     cat.Description,
		ApplicableTypes: cat.ApplicableTypes,
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := easyjson.MarshalToWriter(&response, w); err != nil {
		logger.Error(r.Context(), "easyjson marshal error", logrus.Fields{"error": err})
	}
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		logger.Error(r.Context(), "category delete error", logrus.Fields{"error": err, "id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
