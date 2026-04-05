package repository

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/storage"
	"time"
)

type UserRepo struct {
	store *storage.MemoryStore
}

func NewUserRepo(store *storage.MemoryStore) *UserRepo {
	return &UserRepo{store: store}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	user.ID = r.store.NextUserID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.store.Users[user.ID] = *user
	r.store.UsersByEmail[user.Login] = user.ID
	r.store.UsersByNickname[user.Nickname] = user.ID
	r.store.NextUserID++
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	id, ok := r.store.UsersByEmail[email]
	if !ok {
		return nil, nil
	}
	user := r.store.Users[id]
	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	user, ok := r.store.Users[id]
	if !ok {
		return nil, nil
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	existing, ok := r.store.Users[user.ID]
	if !ok {
		return nil
	}
	existing.Nickname = user.Nickname
	existing.AvatarURL = user.AvatarURL
	existing.UpdatedAt = time.Now()
	r.store.Users[user.ID] = existing
	return nil
}
