package handlers

import (
	"log"
	"net/http"
	"strconv"

	"guidely-app/internal/dto"
	"guidely-app/internal/service"

	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
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
	response := make(dto.CountryResponseList, len(countries))
	for i, c := range countries {
		response[i] = dto.CountryResponse{
			ID:        c.ID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	data, err := easyjson.Marshal(response)
	if err != nil {
		log.Printf("easyjson marshal error: %v", err)
		return
	}
	w.Write(data)
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

	resp := dto.CountryWithLocalitiesResponse{
		ID:         country.ID,
		Name:       country.Name,
		Localities: make([]dto.LocalityResponse, 0, len(localities)),
	}
	for _, l := range localities {
		resp.Localities = append(resp.Localities, dto.LocalityResponse{
			ID:        l.ID,
			Name:      l.Name,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := easyjson.MarshalToWriter(&resp, w); err != nil {
		log.Printf("easyjson marshal error: %v", err)
	}
}
