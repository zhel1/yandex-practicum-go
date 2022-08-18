// Package storage provides interfaces for database.
package storage

import (
	"context"
)

//Users struct
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
	Get(ctx context.Context, key string) (string, error)
	GetUserLinks(ctx context.Context, userID string) (map[string]string, error)
	Put(ctx context.Context, userID, shortURL, originURL string) error
	PutBatch(ctx context.Context, userID string, batchForDB map[string]string) error
	Delete(ctx context.Context, shortURLs []string, userID string) error
	Close() error
}
