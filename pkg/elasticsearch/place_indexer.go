package elasticsearch

import (
	"context"
	"fmt"
	"guidely-app/pkg/models"
)

const PlaceIndex = "places"

type PlaceDocument struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Country     string   `json:"country"`
	Locality    string   `json:"locality"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

var placeIndexMapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"id":          map[string]any{"type": "long"},
			"name":        map[string]any{"type": "text", "analyzer": "standard"},
			"description": map[string]any{"type": "text", "analyzer": "standard"},
			"country":     map[string]any{"type": "text", "analyzer": "standard"},
			"locality":    map[string]any{"type": "text", "analyzer": "standard"},
			"latitude":    map[string]any{"type": "double"},
			"longitude":   map[string]any{"type": "double"},
		},
	},
}

type PlaceIndexer struct {
	client *Client
}

func NewPlaceIndexer(client *Client) *PlaceIndexer {
	return &PlaceIndexer{client: client}
}

func (i *PlaceIndexer) EnsureIndex(ctx context.Context) error {
	return i.client.EnsureIndex(ctx, PlaceIndex, placeIndexMapping)
}

func (i *PlaceIndexer) IndexPlace(ctx context.Context, p models.Place) error {
	doc := PlaceDocument{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Country:     p.Locality.Country,
		Locality:    p.Locality.Name,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
	}
	return i.client.Index(ctx, PlaceIndex, fmt.Sprintf("%d", p.ID), doc)
}

func (i *PlaceIndexer) DeletePlace(ctx context.Context, id uint64) error {
	return i.client.Delete(ctx, PlaceIndex, fmt.Sprintf("%d", id))
}
