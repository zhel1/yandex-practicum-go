// Package http provides handler functions to be used for endpoints for http protocol.
package http

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/http/middleware"
	v1 "github.com/zhel1/yandex-practicum-go/internal/http/v1"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"io"
	"net/http"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.GzipHandler)
	router.Use(middleware.NewCookieHandler(h.services).CookieHandler)

	router.Post("/", h.AddLink())
	router.Get("/{id}", h.GetLink())
	router.Get("/ping", h.Ping())

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router chi.Router) {
	handlerV1 := v1.NewHandler(h.services)
	router.Route("/api", func(r chi.Router) {
		handlerV1.Init(r)
	})
}

// AddLink accepts a URL string in the request body for shortening.
func (h *Handler) AddLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.TakeUserID(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		longLinkBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		status := http.StatusCreated
		shortLink, err := h.services.Shorten.ShortenURL(r.Context(), userID, dto.ModelOriginalURL{
			OriginalURL: string(longLinkBytes),
		})
		if err != nil {
			switch {
			case errors.Is(err, dto.ErrAlreadyExists):
				status = http.StatusConflict
			default:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprint(w, shortLink.ShortURL)
	}
}

//GetLink accepts the identifier of the short URL as a URL parameter and returns a response
func (h *Handler) GetLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		originalLink, err := h.services.Users.GetOriginalURLByShort(r.Context(), shortURL)
		if err != nil {
			switch err {
			case dto.ErrDeleted:
				http.Error(w, err.Error(), http.StatusGone)
			default:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		w.Header().Set("Location", originalLink)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *Handler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.services.Users.Ping(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
