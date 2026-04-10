package handlers

import (
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"guidely-app/internal/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PlaceHandler struct {
	placeService *service.PlaceService
}

func NewPlaceHandler(placeService *service.PlaceService) *PlaceHandler {
	return &PlaceHandler{placeService: placeService}
}

func (h *PlaceHandler) List(w http.ResponseWriter, r *http.Request) {
	places, err := h.placeService.GetAll(r.Context())
	if err != nil {
		utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		return
	}
	response := make([]dto.PlaceResponse, 0, len(places))
	for _, p := range places {
		// упрощённое преобразование (полное опущено для краткости)
		resp := dto.PlaceResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			IsLiked:     false,
			Locality: dto.LocalityDTO{
				ID:        p.Locality.ID,
				Name:      p.Locality.Name,
				Country:   p.Locality.Country,
				Latitude:  p.Locality.Latitude,
				Longitude: p.Locality.Longitude,
			},
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
		if p.Category.ID != 0 {
			resp.Category = &dto.CategoryDTO{
				ID:          p.Category.ID,
				Name:        p.Category.Name,
				Description: p.Category.Description,
			}
		}
		if len(p.Photos) > 0 {
			resp.Photos = make([]dto.PlacePhotoDTO, len(p.Photos))
			for i, ph := range p.Photos {
				resp.Photos[i] = dto.PlacePhotoDTO{
					ID:       ph.ID,
					PlaceID:  ph.PlaceID,
					FilePath: ph.FilePath,
					IsMain:   ph.IsMain,
				}
			}
		}
		response = append(response, resp)
	}
	utils.WriteJSON(w, response, http.StatusOK)
}

func (h *PlaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	place, err := h.placeService.GetByID(r.Context(), id)
	if err != nil || place == nil {
		utils.WriteJSONError(w, utils.ErrNotFound, http.StatusNotFound)
		return
	}
	response := dto.PlaceResponse{
		ID:          place.ID,
		Name:        place.Name,
		Description: place.Description,
		Price:       place.Price,
		IsLiked:     false,
		Locality: dto.LocalityDTO{
			ID:        place.Locality.ID,
			Name:      place.Locality.Name,
			Country:   place.Locality.Country,
			Latitude:  place.Locality.Latitude,
			Longitude: place.Locality.Longitude,
		},
		CreatedAt: place.CreatedAt,
		UpdatedAt: place.UpdatedAt,
	}
	if place.Category.ID != 0 {
		response.Category = &dto.CategoryDTO{
			ID:          place.Category.ID,
			Name:        place.Category.Name,
			Description: place.Category.Description,
		}
	}
	if len(place.Photos) > 0 {
		response.Photos = make([]dto.PlacePhotoDTO, len(place.Photos))
		for i, ph := range place.Photos {
			response.Photos[i] = dto.PlacePhotoDTO{
				ID:       ph.ID,
				PlaceID:  ph.PlaceID,
				FilePath: ph.FilePath,
				IsMain:   ph.IsMain,
			}
		}
	}
	utils.WriteJSON(w, response, http.StatusOK)
}
