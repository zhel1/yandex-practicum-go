package storage

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
//**********************************************************************************************************************
type UserData struct {
	Id string				`json:"id"`
	URLs map[string]string	`json:"urls"`
}

func NewUserData(id string) UserData {
	return UserData{
		Id: id,
		URLs: make(map[string]string),
	}
}
//**********************************************************************************************************************
type Storage interface {
	Get(key string) (string, error)
	GetUserLinks(id string) (map[string]string, error)
	Put(id, key, value string) error
	Close() error
}
