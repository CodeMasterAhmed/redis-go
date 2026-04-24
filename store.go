package main

import (
	"sync"
	"time"
)

type storeEntry struct {
	value     string
	expiresAt time.Time
}

type Store struct {
	mu   sync.RWMutex
	data map[string]storeEntry
	now  func() time.Time
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]storeEntry),
		now:  time.Now,
	}
}

func (s *Store) Set(key, value string) {
	s.SetWithTTL(key, value, 0)
}

func (s *Store) SetWithTTL(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := storeEntry{value: value}
	if ttl > 0 {
		entry.expiresAt = s.now().Add(ttl)
	}
	s.data[key] = entry
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.getLiveEntry(key)
	if !ok {
		return "", false
	}
	return entry.value, true
}

func (s *Store) Delete(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0
	for _, key := range keys {
		if _, ok := s.getLiveEntry(key); ok {
			delete(s.data, key)
			deleted++
		}
	}
	return deleted
}

func (s *Store) Exists(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := 0
	for _, key := range keys {
		if _, ok := s.getLiveEntry(key); ok {
			found++
		}
	}
	return found
}

func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.getLiveEntry(key)
	if !ok {
		return false
	}

	entry.expiresAt = s.now().Add(ttl)
	s.data[key] = entry
	return true
}

func (s *Store) Persist(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.getLiveEntry(key)
	if !ok || entry.expiresAt.IsZero() {
		return false
	}

	entry.expiresAt = time.Time{}
	s.data[key] = entry
	return true
}

func (s *Store) TTL(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.getLiveEntry(key)
	if !ok {
		return -2
	}
	if entry.expiresAt.IsZero() {
		return -1
	}

	ttl := entry.expiresAt.Sub(s.now())
	if ttl <= 0 {
		delete(s.data, key)
		return -2
	}
	return int(ttl / time.Second)
}

func (s *Store) getLiveEntry(key string) (storeEntry, bool) {
	entry, ok := s.data[key]
	if !ok {
		return storeEntry{}, false
	}
	if !entry.expiresAt.IsZero() && !entry.expiresAt.After(s.now()) {
		delete(s.data, key)
		return storeEntry{}, false
	}
	return entry, true
}
