package storage

import (
	"encoding/json"
	"log"
	"os"
)

type InFile struct {
	file *os.File
	cache Storage
	encoder *json.Encoder
}

func NewInFile(fileName string) (Storage, error){
	file, err := os.OpenFile(fileName, os.O_RDWR | os.O_CREATE, 0755)
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

	return &InFile {
		file: file,
		cache: data,
		encoder: json.NewEncoder(file),
	}, nil
}

func (s *InFile) Get(key string) (string, error) {
	return s.cache.Get(key)
}

func (s *InFile) GetUserLinks(id string) (map[string]string, error){
	return s.cache.GetUserLinks(id)
}

func (s *InFile) Put(id string, key, value string) error {
	if err := s.cache.Put(id, key, value); err != nil {
		return err
	}

	//rewrite all file
	s.file.Truncate(0)
	s.file.Seek(0,0)
	return s.encoder.Encode(&s.cache)
}

func (s *InFile) Close() error {
	s.cache = nil
	return s.file.Close()
}