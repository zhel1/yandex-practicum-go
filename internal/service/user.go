// Package service implements the business logic of the application.
package service

import (
	"context"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
)

// Check interface implementation
var (
	_ User = (*UserService)(nil)
)

type UserService struct {
	storage storage.Storage
	baseURL string
}

func NewUserService(storage storage.Storage, baseURL string) *UserService {
	return &UserService{
		storage: storage,
		baseURL: baseURL,
	}
}

func (s *UserService) GetOriginalURLByShort(ctx context.Context, shortURL string) (string, error) {
	return s.storage.Get(ctx, shortURL)
}

func (s *UserService) GetURLsByUserID(ctx context.Context, UserID string) ([]dto.ModelURL, error) {
	links, err := s.storage.GetUserLinks(ctx, UserID)
	if err != nil {
		return nil, err
	}

	responseURLs := make([]dto.ModelURL, 0, len(links))
	for short, orign := range links {
		responseURL := dto.ModelURL{
			OriginalURL: orign,
			ShortURL:    s.baseURL + short,
		}
		responseURLs = append(responseURLs, responseURL)
	}

	return responseURLs, nil
}

func (s *UserService) DeleteBatchURL(ctx context.Context, userID string, shortURLs []string) error {
	// perform asynchronous deletion
	return s.storage.Delete(ctx, shortURLs, userID)
}

func (s *UserService) Ping(ctx context.Context) error {
	pinger, valid := s.storage.(storage.Pinger)
	if valid {
		if err := pinger.PingDB(); err != nil {
			return err
		}
		return nil
	}
	panic("Storage doesn't implement interface Pinger")
}
