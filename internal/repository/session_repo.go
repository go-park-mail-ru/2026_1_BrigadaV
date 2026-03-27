package repository

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/storage"
	"time"
)

type SessionRepo struct {
	store *storage.MemoryStore
}

func NewSessionRepo(store *storage.MemoryStore) *SessionRepo {
	return &SessionRepo{store: store}
}

func (r *SessionRepo) Create(ctx context.Context, session *models.Session) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	session.ID = uint64(len(r.store.Sessions) + 1)
	session.CreatedAt = time.Now()
	r.store.Sessions[session.SessionToken] = *session
	return nil
}

func (r *SessionRepo) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	session, ok := r.store.Sessions[token]
	if !ok {
		return nil, nil
	}
	return &session, nil
}

func (r *SessionRepo) DeleteByToken(ctx context.Context, token string) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	delete(r.store.Sessions, token)
	return nil
}
