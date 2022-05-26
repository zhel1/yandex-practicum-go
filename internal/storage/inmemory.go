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

func (s *InMemory) Get(key string) (string, error){
	s.RLock()
	defer s.RUnlock()
	for _, usrData := range s.m {
		if v, ok := usrData.URLs[key]; ok {
			return v, nil
		}
	}
	return "", ErrNotFound
}
func (s *InMemory) GetUserLinks(id string) (map[string]string, error){
	s.RLock()
	defer s.RUnlock()
	if usrData, ok := s.m[id]; ok {
		return usrData.URLs, nil
	} else {
		return nil, ErrNotFound
	}
}

func (s *InMemory) Put(id, key, value string) error {
	s.Lock()
	defer s.Unlock()
	if usrData, ok := s.m[id]; ok {
		//User with uuid exists.
		//Just append new item to URLs
		if _, ok := usrData.URLs[key]; ok {
			return ErrAlreadyExists
		}
		usrData.URLs[key] = value
	} else {
		usrData := NewUserData(id)
		usrData.URLs[key] = value
		s.m[id] = usrData
	}
	return nil
}

func (s *InMemory) Close() error {
	s.Lock()
	defer s.Unlock()
	s.m = nil
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