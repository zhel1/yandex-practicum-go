package service

import (
	"context"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
)

// Check interface implementation
var (
	_ Internal = (*InternalService)(nil)
)

type InternalService struct {
	storage storage.Storage
}

func NewInternalService(storage storage.Storage) *InternalService {
	return &InternalService{
		storage: storage,
	}
}

func (s *InternalService) GetURLsCount(ctx context.Context) (int, error) {
	return s.storage.GetURLsCount(ctx)
}

func (s *InternalService) GetUsersCount(ctx context.Context) (int, error) {
	return s.storage.GetUserCount(ctx)
}

func (s *InternalService) GetStatistic(ctx context.Context) (dto.Statistic, error) {
	urlsCount, err := s.GetURLsCount(ctx)
	if err != nil {
		return dto.Statistic{}, err
	}

	usersCount, err := s.GetUsersCount(ctx)
	if err != nil {
		return dto.Statistic{}, err
	}

	return dto.Statistic{
		URLsCount:  urlsCount,
		UsersCount: usersCount,
	}, nil
}
