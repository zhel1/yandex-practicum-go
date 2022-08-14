package storage

import (
	"encoding/json"
	"log"
	"os"
)

// DB in file struct
type InFile struct {
	file    *os.File
	cache   Storage
	encoder *json.Encoder
}

// NewInFile is DB constructor
func NewInFile(fileName string) (Storage, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	data := NewInMemory()

	if stat, _ := file.Stat(); stat.Size() != 0 {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&data); err != nil {
			log.Fatal("DB file is damaged.")
			return nil, err
		}
	}

	return &InFile{
		file:    file,
		cache:   data,
		encoder: json.NewEncoder(file),
	}, nil
}

// Get gets base URL from DB
func (s *InFile) Get(shortURL string) (string, error) {
	return s.cache.Get(shortURL)
}

// GetUserLinks gets all URLs by UserID from DB
func (s *InFile) GetUserLinks(userID string) (map[string]string, error) {
	return s.cache.GetUserLinks(userID)
}

// Put sets short URL in DB
func (s *InFile) Put(userID string, shortURL, originURL string) error {
	if err := s.cache.Put(userID, shortURL, originURL); err != nil {
		return err
	}

	//rewrite all file
	s.file.Truncate(0)
	s.file.Seek(0, 0)
	return s.encoder.Encode(&s.cache)
}

// Close removes cache and close thr file
func (s *InFile) Close() error {
	s.cache = nil
	return s.file.Close()
}

// Delete does nothing
// It is necessary for the implementation of the interface Storage
func (s *InFile) Delete(shortURLs []string, userID string) error {
	return nil
}
