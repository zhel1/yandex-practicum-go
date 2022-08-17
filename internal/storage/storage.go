package storage

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrDeleted       = errors.New("marked as deleted")
	ErrAlreadyExists = errors.New("already exists")
	ErrExecutionPSQL = errors.New("execution PSQL error")
	ErrStatementPSQL = errors.New("statement PSQL error")
)

//**********************************************************************************************************************

//Users interface
type UserData struct {
	ID   string            `json:"id"`
	URLs map[string]string `json:"urls"`
}

//UserData constructor
func NewUserData(id string) UserData {
	return UserData{
		ID:   id,
		URLs: make(map[string]string),
	}
}

//**********************************************************************************************************************

//Pinger interface
type Pinger interface {
	PingDB() error
}

//**********************************************************************************************************************

//Storage interface
type Storage interface {
	Get(key string) (string, error)
	GetUserLinks(userID string) (map[string]string, error)
	Put(userID, shortURL, originURL string) error
	Close() error
	Delete(shortURLs []string, userID string) error
}
