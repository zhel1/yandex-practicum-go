package storage

import (
	"encoding/json"
	"sync"
)

//DB in memory struct
type InMemory struct {
	sync.RWMutex
	m map[string]UserData
}

//NewInMemory is DB constructor
func NewInMemory() Storage {
	return &InMemory{
		m: make(map[string]UserData),
	}
}

//Get gets base URL from DB
func (s *InMemory) Get(shortURL string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	for _, usrData := range s.m {
		if v, ok := usrData.URLs[shortURL]; ok {
			return v, nil
		}
	}
	return "", ErrNotFound
}

//GetUserLinks gets all URLs by UserID from DB
func (s *InMemory) GetUserLinks(userID string) (map[string]string, error) {
	s.RLock()
	defer s.RUnlock()
	if usrData, ok := s.m[userID]; ok {
		return usrData.URLs, nil
	} else {
		return nil, ErrNotFound
	}
}

//Put sets short URL in DB
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

//Close clears the map with user data
func (s *InMemory) Close() error {
	s.Lock()
	defer s.Unlock()
	s.m = nil
	return nil
}

//Close stops active workers disconnects from DB
func (s *InMemory) Delete(shortURLs []string, userID string) error {
	return nil
}

//**********************************************************************************************************************

//MarshalJSON serializes the database given in json format
func (s *InMemory) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.m)
}

//UnmarshalJSON deserializes the database given from json format
func (s *InMemory) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &((*s).m)); err != nil {
		return err
	}
	return nil
}
