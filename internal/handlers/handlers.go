package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/middleware"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"io"
	"net/http"
	"net/url"
)

func shortenURL(st storage.Storage, context context.Context, baseURL, URL string) (string, error) {
	userIDCtx := ""
	if id := context.Value(middleware.UserIDCtxName); id != nil {
		userIDCtx = id.(string)
	}

	if userIDCtx == "" {
		return "", errors.New("empty user id")
	}

	if _, err := url.ParseRequestURI(URL); err != nil {
		return "", err
	}

	shortIDLink := utils.MD5(URL)[:8]

	if err := st.Put(userIDCtx, shortIDLink, URL); err != nil {
		return "", err
	}
	return baseURL + shortIDLink, nil
}
//**********************************************************************************************************************
type URLHandler struct {
	st storage.Storage
	config *config.Config
}

func InitURLHandler(storage storage.Storage, config *config.Config) (*URLHandler, error) {
	if storage == nil {
		return nil, fmt.Errorf("nil Storage was passed to service URL Handler initializer")
	}
	if config == nil {
		return nil, fmt.Errorf("nil Config was passed to service URL Handler initializer")
	}
	return &URLHandler{st: storage, config: config}, nil
}

func (h *URLHandler)AddLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		longLinkBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortLink, err := shortenURL(h.st, r.Context(), h.config.BaseURL, string(longLinkBytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortLink)
	}
}

type JSONRequestData struct {
	URL string `json:"url"`
}

type JSONResponsetData struct {
	Result string `json:"result"`
}

func (h *URLHandler)AddLinkJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b := JSONRequestData {}
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shortLink, err := shortenURL(h.st, r.Context(), h.config.BaseURL, b.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res := JSONResponsetData {
			Result: shortLink,
		}

		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		if err = encoder.Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, buf)
	}
}

func (h *URLHandler)GetLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		linkID := chi.URLParam(r, "id")
		longLink, err := h.st.Get(linkID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			w.Header().Set("Location", longLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}

// ResponseFullURL is used in GetUserLinks
type ResponseFullURL struct {
	OriginalURL  string `json:"original_url"`
	ShortURL string `json:"short_url"`
}

func (h *URLHandler)GetUserLinks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userIDCtx string
		if id := r.Context().Value(middleware.UserIDCtxName); id != nil {
			userIDCtx = id.(string)
		}

		if userIDCtx == "" {
			http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		}

		links, err := h.st.GetUserLinks(userIDCtx)
		if err != nil || len(links) == 0  {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		var responseURLs []ResponseFullURL
		for short, orign := range links {
			responseURL := ResponseFullURL{
				OriginalURL: orign,
				ShortURL: h.config.BaseURL + short,
			}
			responseURLs = append(responseURLs, responseURL)
		}

		//TODO remove code duplication (the same piece in AddLinkJSON)
		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		if err = encoder.Encode(responseURLs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf)
	}
}

func (h *URLHandler)GetPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pinger, valid := h.st.(storage.Pinger)
		if valid {
			if err := pinger.PingDB(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}
}