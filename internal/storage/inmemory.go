package storage

type InMemory struct {
	m map[string]string
}

func NewInMemory() Storage {
	return &InMemory{
		m: make(map[string]string),
	}
}

func (s *InMemory) Get(key string) (string, error) {
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return "", ErrNotFound
}

func (s *InMemory) Put(key string, value string) error {
	if _, ok := s.m[key]; ok {
		return ErrAlreadyExists
	}
	s.m[key] = value
	return nil
}

func (s *InMemory) Close() error {
	s.m = nil
	return nil
}
