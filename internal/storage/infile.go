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

func (s *InFile) Get(shortURL string) (string, error) {
	return s.cache.Get(shortURL)
}

func (s *InFile) GetUserLinks(userID string) (map[string]string, error){
	return s.cache.GetUserLinks(userID)
}

func (s *InFile) Put(userID string, shortURL, originURL string) error {
	if err := s.cache.Put(userID, shortURL, originURL); err != nil {
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

// SendToQueue is a mock for PSQL DB batch concurrent deleter.
func (s *InFile) Delete(shortURLs []string, userID string) error {
	return nil
}
