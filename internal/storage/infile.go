package storage

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type InFile struct {
	sync.RWMutex
	file *os.File
	cache map[string]string
	encoder *json.Encoder
}

func NewInFile(fileName string) (Storage, error){
	file, err := os.OpenFile(fileName, os.O_RDWR | os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)

	if stat, _ := file.Stat(); stat.Size() != 0 {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&data); err != nil {
			log.Fatal("DB file is damaged.")
			return nil, err
		}
	}

	return &InFile {
		file:    file,
		cache: data,
		encoder: json.NewEncoder(file),
	}, nil
}

func (s *InFile) Close() error {
	s.Lock()
	defer s.Unlock()
	s.cache = nil
	return s.file.Close()
}

func (s *InFile) Get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	if v, ok := s.cache[key]; ok {
		return v, nil
	}
	return "", ErrNotFound
}

func (s *InFile) Put(key string, value string) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.cache[key]; ok {
		return ErrAlreadyExists
	}
	s.cache[key] = value

	return s.encoder.Encode(&s.cache)
}
