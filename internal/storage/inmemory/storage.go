// Package inmemory implements storage in ram.
package inmemory

import (
	"context"
	"encoding/json"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	storageErrors "github.com/zhel1/yandex-practicum-go/internal/storage/errors"
	"sync"
)

// Check interface implementation.
var (
	_ storage.Storage = (*Storage)(nil)
)

// Storage is DB in memory struct
type Storage struct {
	sync.RWMutex
	m map[string]storage.UserData
}

// NewStorage creates DB in memory.
func NewStorage() storage.Storage {
	return &Storage{
		m: make(map[string]storage.UserData),
	}
}

// Get gets base URL from DB.
func (s *Storage) Get(ctx context.Context, shortURL string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	for _, usrData := range s.m {
		if v, ok := usrData.URLs[shortURL]; ok {
			return v, nil
		}
	}
	return "", &storageErrors.NotFoundError{Err: dto.ErrNotFound}
}

// GetUserLinks returns all URLs by UserID from DB.
func (s *Storage) GetUserLinks(ctx context.Context, userID string) (map[string]string, error) {
	s.RLock()
	defer s.RUnlock()
	if usrData, ok := s.m[userID]; ok {
		return usrData.URLs, nil
	} else {
		return nil, &storageErrors.NotFoundError{Err: dto.ErrNotFound}
	}
}

// Put save short URL in DB.
func (s *Storage) Put(ctx context.Context, userID, shortURL, originURL string) error {
	s.Lock()
	defer s.Unlock()
	if usrData, ok := s.m[userID]; ok {
		//User with uuid exists.
		//Just append new item to URLs
		if _, ok := usrData.URLs[shortURL]; ok {
			return &storageErrors.AlreadyExistsError{Err: dto.ErrAlreadyExists}
		}
		usrData.URLs[shortURL] = originURL
	} else {
		usrData := storage.NewUserData(userID)
		usrData.URLs[shortURL] = originURL
		s.m[userID] = usrData
	}
	return nil
}

// Put save short URLs in DB.
func (s *Storage) PutBatch(ctx context.Context, userID string, batchForDB map[string]string) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.m[userID]; !ok {
		usrData := storage.NewUserData(userID)
		s.m[userID] = usrData
	}

	for originURL, shortURL := range batchForDB {
		if _, ok := s.m[userID].URLs[shortURL]; ok {
			return &storageErrors.AlreadyExistsError{Err: dto.ErrAlreadyExists}
		}
		s.m[userID].URLs[shortURL] = originURL
	}

	return nil
}

// Delete needs to implement Storage interface.
func (s *Storage) Delete(ctx context.Context, shortURLs []string, userID string) error {
	return nil
}

// GetURLsCount returns the mount of all urls in DB
func (s *Storage) GetURLsCount(ctx context.Context) (int, error) {
	urlsCount := 0
	for _, userData := range s.m {
		urlsCount = +len(userData.URLs)
	}
	return urlsCount, nil
}

// GetURLsCount returns the mount of all users in DB
func (s *Storage) GetUserCount(ctx context.Context) (int, error) {
	return len(s.m), nil
}

//PingDB checks connection to DB
func (s *Storage) PingDB() error {
	return nil
}

// Close clears the map with user data.
func (s *Storage) Close() error {
	s.Lock()
	defer s.Unlock()
	s.m = nil
	return nil
}

//**********************************************************************************************************************

//MarshalJSON serializes the database given in json format
func (s *Storage) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.m)
}

//UnmarshalJSON deserializes the database given from json format
func (s *Storage) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &((*s).m)); err != nil {
		return err
	}
	return nil
}
