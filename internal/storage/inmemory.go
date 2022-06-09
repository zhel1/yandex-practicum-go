package storage

import (
	"encoding/json"
	"sync"
)

type InMemory struct {
	sync.RWMutex
	m map[string]UserData
}

func NewInMemory() Storage {
	return &InMemory{
		m:  make(map[string]UserData),
	}
}

func (s *InMemory) Get(shortURL string) (string, error){
	s.RLock()
	defer s.RUnlock()
	for _, usrData := range s.m {
		if v, ok := usrData.URLs[shortURL]; ok {
			return v, nil
		}
	}
	return "", ErrNotFound
}
func (s *InMemory) GetUserLinks(userID string) (map[string]string, error){
	s.RLock()
	defer s.RUnlock()
	if usrData, ok := s.m[userID]; ok {
		return usrData.URLs, nil
	} else {
		return nil, ErrNotFound
	}
}

func (s *InMemory) Put(userID, shortURL, originURL string) error {
	s.Lock()
	defer s.Unlock()
	if usrData, ok := s.m[userID]; ok {
		//User with uuid exists.
		//Just append new item to URLs
		if _, ok := usrData.URLs[shortURL]; ok {
			return ErrAlreadyExists
		}
		usrData.URLs[shortURL] = originURL
	} else {
		usrData := NewUserData(userID)
		usrData.URLs[shortURL] = originURL
		s.m[userID] = usrData
	}
	return nil
}

func (s *InMemory) Close() error {
	s.Lock()
	defer s.Unlock()
	s.m = nil
	return nil
}

// SendToQueue is a mock for PSQL DB batch concurrent deleter.
func (s *InMemory) Delete(shortURLs []string, userID string) error {
	return nil
}
//**********************************************************************************************************************
func (s *InMemory) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.m)
}

func (s *InMemory) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &((*s).m)); err != nil {
		return err
	}
	return nil
}