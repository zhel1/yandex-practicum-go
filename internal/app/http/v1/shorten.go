package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/app/dto"
	"github.com/zhel1/yandex-practicum-go/internal/app/http/middleware"
	"net/http"
)

func (h *Handler) initShortenRoutes(r chi.Router) {
	r.Route("/shorten", func(r chi.Router) {
		r.Post("/", h.AddLinkJSON())
		r.Post("/batch", h.AddLinkBatchJSON())
	})
}

//AddLinkJSON accepts a JSON object in the request body and returning a JSON object in response
func (h *Handler) AddLinkJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.TakeUserID(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b := dto.ModelOriginalURL{}
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		modelShortURL, err := h.services.Shorten.ShortenURL(r.Context(), userID, b)
		if err != nil && !errors.Is(err, dto.ErrAlreadyExists) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		status := http.StatusCreated
		if errors.Is(err, dto.ErrAlreadyExists) {
			status = http.StatusConflict
		}

		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		if err = encoder.Encode(modelShortURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprint(w, buf)
	}
}

//PostURLsBATCH accepts in the request body a set of URLs for shortening in the format
func (h *Handler) AddLinkBatchJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.TakeUserID(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		bReq := make([]dto.ModelOriginalURLBatch, 0, 5)
		if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		bResArr, err := h.services.Shorten.ShortenBatchURL(r.Context(), userID, bReq)
		if errors.Is(err, dto.ErrAlreadyExists) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(bResArr); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, buf)
	}
}
