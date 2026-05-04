package storage

import (
	"guidely-app/pkg/models"
	"sync"
)

type MemoryStore struct {
	Users           map[uint64]models.User
	UsersByEmail    map[string]uint64
	UsersByNickname map[string]uint64
	Sessions        map[string]models.Session
	Places          map[uint64]models.Place
	UserLikes       map[uint64]map[uint64]bool
	NextUserID      uint64
	Mu              sync.RWMutex
	LikesMu         sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Users:           make(map[uint64]models.User),
		UsersByEmail:    make(map[string]uint64),
		UsersByNickname: make(map[string]uint64),
		Sessions:        make(map[string]models.Session),
		Places:          make(map[uint64]models.Place),
		UserLikes:       make(map[uint64]map[uint64]bool),
		NextUserID:      1,
	}
}
