package main

import "sync"

// Store is the interface for all storage engine implementations.
// Each stage of the database build swaps the implementation while
// keeping the HTTP API unchanged.
type Store interface {
	Get(key string) ([]byte, bool)
	Put(key string, value []byte)
	Delete(key string) bool
	Len() int
}

// MemStore is Stage 1: an in-memory key-value store backed by a
// Go map protected with a read-write mutex.
type MemStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemStore() *MemStore {
	return &MemStore{data: make(map[string][]byte)}
}

func (s *MemStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	return v, ok
}

func (s *MemStore) Put(key string, value []byte) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func (s *MemStore) Delete(key string) bool {
	s.mu.Lock()
	_, ok := s.data[key]
	if ok {
		delete(s.data, key)
	}
	s.mu.Unlock()
	return ok
}

func (s *MemStore) Len() int {
	s.mu.RLock()
	n := len(s.data)
	s.mu.RUnlock()
	return n
}
