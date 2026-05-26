package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"guidely-app/internal/service"

	"github.com/gorilla/mux"
)

type CountryHandler struct {
	svc service.CountryService
}

func NewCountryHandler(svc service.CountryService) *CountryHandler {
	return &CountryHandler{svc: svc}
}

func (h *CountryHandler) List(w http.ResponseWriter, r *http.Request) {
	countries, err := h.svc.GetAll(r.Context())
	if err != nil {
		log.Printf("country list error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(countries)
}

type countryWithLocalitiesResponse struct {
	ID        uint64              `json:"id"`
	Name      string              `json:"name"`
	Localities []localityResponse `json:"localities"`
}

type localityResponse struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

func (h *CountryHandler) GetWithLocalities(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid country id", http.StatusBadRequest)
		return
	}

	country, localities, err := h.svc.GetWithLocalities(r.Context(), id)
	if err != nil {
		log.Printf("country get with localities error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if country == nil {
		http.Error(w, "country not found", http.StatusNotFound)
		return
	}

	resp := countryWithLocalitiesResponse{
		ID:         country.ID,
		Name:       country.Name,
		Localities: make([]localityResponse, 0, len(localities)),
	}
	for _, l := range localities {
		resp.Localities = append(resp.Localities, localityResponse{
			ID:        l.ID,
			Name:      l.Name,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
