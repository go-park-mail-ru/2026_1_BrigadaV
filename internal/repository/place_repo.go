package repository

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/storage"
)

type PlaceRepo struct {
	store *storage.MemoryStore
}

func NewPlaceRepo(store *storage.MemoryStore) *PlaceRepo {
	return &PlaceRepo{store: store}
}

func (r *PlaceRepo) GetAll(ctx context.Context) ([]models.Place, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	places := make([]models.Place, 0, len(r.store.Places))
	for _, p := range r.store.Places {
		places = append(places, p)
	}
	return places, nil
}

func (r *PlaceRepo) GetByID(ctx context.Context, id uint64) (*models.Place, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	p, ok := r.store.Places[id]
	if !ok {
		return nil, nil
	}
	return &p, nil
}
