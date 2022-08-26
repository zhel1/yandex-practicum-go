// Package infile implements storage in file.
package infile

import (
	"context"
	"encoding/json"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inmemory"
	"log"
	"os"
)

//Check interface implementation
var (
	_ storage.Storage = (*Storage)(nil)
)

// Storage is DB in file struct
type Storage struct {
	file    *os.File
	cache   storage.Storage
	encoder *json.Encoder
}

// NewStorage is DB constructor
func NewStorage(fileName string) (storage.Storage, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	data := inmemory.NewStorage()

	if stat, _ := file.Stat(); stat.Size() != 0 {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&data); err != nil {
			log.Fatal("DB file is damaged.")
			return nil, err
		}
	}

	return &Storage{
		file:    file,
		cache:   data,
		encoder: json.NewEncoder(file),
	}, nil
}

// Get gets base URL from DB
func (s *Storage) Get(ctx context.Context, shortURL string) (string, error) {
	return s.cache.Get(ctx, shortURL)
}

// GetUserLinks gets all URLs by UserID from DB
func (s *Storage) GetUserLinks(ctx context.Context, userID string) (map[string]string, error) {
	return s.cache.GetUserLinks(ctx, userID)
}

// Put sets short URL in DB
func (s *Storage) Put(ctx context.Context, userID string, shortURL, originURL string) error {
	if err := s.cache.Put(ctx, userID, shortURL, originURL); err != nil {
		return err
	}

	//rewrite all file
	s.file.Truncate(0)
	s.file.Seek(0, 0)
	return s.encoder.Encode(&s.cache)
}

func (s *Storage) PutBatch(ctx context.Context, userID string, batchForDB map[string]string) error {
	s.cache.PutBatch(ctx, userID, batchForDB)

	//rewrite all file
	s.file.Truncate(0)
	s.file.Seek(0, 0)
	return s.encoder.Encode(&s.cache)
}

// Delete does nothing
// It is necessary for the implementation of the interface Storage
func (s *Storage) Delete(ctx context.Context, shortURLs []string, userID string) error {
	return nil
}

// Close removes cache and close thr file
func (s *Storage) Close() error {
	s.cache = nil
	return s.file.Close()
}
