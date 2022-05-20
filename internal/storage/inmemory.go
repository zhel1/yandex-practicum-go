package storage

import "sync"

type InMemory struct {
	sync.RWMutex
	m map[string]string
}

func NewInMemory() Storage {
	return &InMemory{
		m: make(map[string]string),
	}
}

func (s *InMemory) Get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return "", ErrNotFound
}

func (s *InMemory) Put(key string, value string) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.m[key]; ok {
		return ErrAlreadyExists
	}
	s.m[key] = value
	return nil
}

func (s *InMemory) Close() error {
	s.Lock()
	defer s.Unlock()
	s.m = nil
	return nil
}
