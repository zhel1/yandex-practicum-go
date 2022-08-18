// Package service implements the business logic of the application.
package service

import (
	"context"
	"errors"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"net/url"
)

// Check interface implementation
var (
	_ Shorten = (*ShortenService)(nil)
)

type ShortenService struct {
	baseURL string
	storage storage.Storage
}

func NewShortenService(storage storage.Storage, baseURL string) *ShortenService {
	return &ShortenService{
		baseURL: baseURL,
		storage: storage,
	}
}

func (s *ShortenService) ShortenURL(ctx context.Context, userID string, URL dto.ModelOriginalURL) (dto.ModelShortURL, error) {
	if _, err := url.ParseRequestURI(URL.OriginalURL); err != nil {
		return dto.ModelShortURL{}, err
	}

	shortIDLink := utils.MD5(URL.OriginalURL)[:8]

	response := dto.ModelShortURL{
		ShortURL: s.baseURL + shortIDLink,
	}

	if err := s.storage.Put(ctx, userID, shortIDLink, URL.OriginalURL); err != nil {
		switch {
		case errors.Is(err, dto.ErrAlreadyExists):
			return response, err
		default:
			return dto.ModelShortURL{}, err
		}
	}
	return response, nil
}

func (s *ShortenService) ShortenBatchURL(ctx context.Context, userID string, URLs []dto.ModelOriginalURLBatch) ([]dto.ModelShortURLBatch, error) {
	bResArr := make([]dto.ModelShortURLBatch, 0, len(URLs))
	batchForDB := make(map[string]string, len(URLs))

	for _, batch := range URLs {
		short := utils.MD5(batch.OriginalURL)[:8]

		bRes := dto.ModelShortURLBatch{
			CorrelationID: batch.CorrelationID,
			ShortURL:      s.baseURL + short,
		}

		batchForDB[batch.OriginalURL] = short

		bResArr = append(bResArr, bRes)
	}

	if err := s.storage.PutBatch(ctx, userID, batchForDB); err != nil {
		switch {
		case errors.Is(err, dto.ErrAlreadyExists):
			return nil, dto.ErrAlreadyExists
		default:
			return nil, err
		}
	}
	return bResArr, nil
}
