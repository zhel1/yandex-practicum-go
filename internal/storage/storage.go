package storage

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrExecutionPSQL = errors.New("execution PSQL error")
)
//**********************************************************************************************************************
type UserData struct {
	ID   string            `json:"id"`
	URLs map[string]string `json:"urls"`
}

func NewUserData(id string) UserData {
	return UserData{
		ID:   id,
		URLs: make(map[string]string),
	}
}
//**********************************************************************************************************************
type Pinger interface {
	PingDB() error
}
//**********************************************************************************************************************
type Storage interface {
	Get(key string) (string, error)
	GetUserLinks(userID string) (map[string]string, error)
	Put(userID, linkID, originURL string) error
	Close() error
}
