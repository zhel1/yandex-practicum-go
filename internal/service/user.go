// Package service implements the business logic of the application.
package service

import (
	"context"
	"github.com/zhel1/yandex-practicum-go/internal/auth"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"time"
)

// Check interface implementation
var (
	_ User = (*UserService)(nil)
)

type UserService struct {
	storage      storage.Storage
	baseURL      string
	tokenManager auth.TokenManager
}

func NewUserService(storage storage.Storage, baseURL string, tokenManager auth.TokenManager) *UserService {
	return &UserService{
		storage:      storage,
		baseURL:      baseURL,
		tokenManager: tokenManager,
	}
}

func (s *UserService) CreateNewToken(ctx context.Context, userID string) (string, error) {
	var expiredAt time.Duration = 1<<63 - 1 //forever
	token, err := s.tokenManager.NewJWT(userID, expiredAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) CheckToken(ctx context.Context, token string) (string, error) {
	userID, err := s.tokenManager.Parse(token)
	if err != nil {
		return "", err
	}
	return userID, nil
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
