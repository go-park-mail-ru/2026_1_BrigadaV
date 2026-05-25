package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"guidely-app/internal/logger"
	"guidely-app/internal/service"
	"guidely-app/pkg/models"

	"github.com/gorilla/mux"
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
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		logger.Error(r.Context(), "category get error", logrus.Fields{"error": err, "id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if c == nil {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var cat models.Category
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.svc.Create(r.Context(), &cat); err != nil {
		logger.Error(r.Context(), "category create error", logrus.Fields{"error": err})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cat)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}
	var cat models.Category
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	cat.ID = id
	if err := h.svc.Update(r.Context(), &cat); err != nil {
		logger.Error(r.Context(), "category update error", logrus.Fields{"error": err, "id": id})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cat)
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
