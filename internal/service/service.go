// Package service implements the business logic of the application.
package service

import (
	"context"
	"github.com/zhel1/yandex-practicum-go/internal/auth"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"net"
)

type User interface {
	CreateNewToken(ctx context.Context, userID string) (string, error)
	CheckToken(ctx context.Context, token string) (string, error)
	GetOriginalURLByShort(ctx context.Context, shortURL string) (string, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]dto.ModelURL, error)
	DeleteBatchURL(ctx context.Context, userID string, shortURLs []string) error
	Ping(ctx context.Context) error
}

type Shorten interface {
	ShortenURL(ctx context.Context, userID string, URL dto.ModelOriginalURL) (dto.ModelShortURL, error)
	ShortenBatchURL(ctx context.Context, userID string, URLs []dto.ModelOriginalURLBatch) ([]dto.ModelShortURLBatch, error)
}

type Internal interface {
	GetURLsCount(ctx context.Context) (int, error)
	GetUsersCount(ctx context.Context) (int, error)

	GetStatistic(ctx context.Context) (dto.Statistic, error)
}

type Security interface {
	IsIpAddrTrusted(ctx context.Context, ipStr string) (bool, error)
}

type Services struct {
	Users    User
	Shorten  Shorten
	Internal Internal
	Security Security
}

type Deps struct {
	Storage       storage.Storage
	BaseURL       string
	TokenManager  auth.TokenManager
	TrustedSubnet *net.IPNet
}

func NewServices(deps Deps) *Services {
	return &Services{
		Shorten:  NewShortenService(deps.Storage, deps.BaseURL),
		Users:    NewUserService(deps.Storage, deps.BaseURL, deps.TokenManager),
		Internal: NewInternalService(deps.Storage),
		Security: NewSecurityService(deps.TrustedSubnet),
	}
}
