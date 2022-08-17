package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/app/service"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Init(r chi.Router) {
	h.initShortenRoutes(r)
	h.initUserRoutes(r)
}
