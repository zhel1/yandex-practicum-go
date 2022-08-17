package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/middleware"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
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
		if errors.Is(err, storage.ErrAlreadyExists) {
			return baseURL + shortIDLink, err
		}
	}
	return baseURL + shortIDLink, nil
}

//**********************************************************************************************************************

//Handler struct
type URLHandler struct {
	st     storage.Storage
	config *config.Config
}

//InitURLHandler initializes Handler struct
func InitURLHandler(storage storage.Storage, config *config.Config) (*URLHandler, error) {
	if storage == nil {
		return nil, fmt.Errorf("nil Storage was passed to service URL Handler initializer")
	}
	if config == nil {
		return nil, fmt.Errorf("nil Config was passed to service URL Handler initializer")
	}
	return &URLHandler{st: storage, config: config}, nil
}

//AddLink accepts a URL string in the request body for shortening
func (h *URLHandler) AddLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		longLinkBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortLink, err := shortenURL(h.st, r.Context(), h.config.BaseURL, string(longLinkBytes))
		if err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		status := http.StatusCreated
		if errors.Is(err, storage.ErrAlreadyExists) {
			status = http.StatusConflict
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprint(w, shortLink)
	}
}

//TODO move into separate file

//JSONRequestData struct
type JSONRequestData struct {
	URL string `json:"url"`
}

//JSONResponsetData struct
type JSONResponsetData struct {
	Result string `json:"result"`
}

//AddLinkJSON accepts a JSON object in the request body and returning a JSON objec in response
func (h *URLHandler) AddLinkJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b := JSONRequestData{}
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shortLink, err := shortenURL(h.st, r.Context(), h.config.BaseURL, b.URL)
		if err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		status := http.StatusCreated
		if errors.Is(err, storage.ErrAlreadyExists) {
			status = http.StatusConflict
		}

		res := JSONResponsetData{
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
		w.WriteHeader(status)
		fmt.Fprint(w, buf)
	}
}

//TODO move into separate file

//JSONResponsetData struct
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//BatchResponse struct
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//PostURLsBATCH accepts in the request body a set of URLs for shortening in the format
func (h *URLHandler) AddLinkBatchJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bReq := make([]BatchRequest, 0, 20000)
		if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		bResArr := make([]BatchResponse, 0, 20000)

		//TODO Writing to the database using statements
		for _, batch := range bReq {
			bRes := BatchResponse{}
			var err error
			bRes.CorrelationID = batch.CorrelationID
			bRes.ShortURL, err = shortenURL(h.st, r.Context(), h.config.BaseURL, batch.OriginalURL)

			if errors.Is(err, storage.ErrAlreadyExists) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			bResArr = append(bResArr, bRes)
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

//GetLink accepts the identifier of the short URL as a URL parameter and returns a response
func (h *URLHandler) GetLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		longLink, err := h.st.Get(shortURL)

		if err != nil {
			switch err {
			case storage.ErrDeleted:
				log.Println("GetLink... StatusGone ", shortURL)
				http.Error(w, err.Error(), http.StatusGone)
			default:
				log.Println("GetLink... StatusBadRequest ", shortURL)
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		} else {
			log.Println("GetLink... StatusTemporaryRedirect ", shortURL)
			w.Header().Set("Location", longLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}

// ResponseFullURL is used in GetUserLinks
type ResponseFullURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

//GetUserLinks returns to the user all links ever saved by him
func (h *URLHandler) GetUserLinks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDCtx := ""
		if id := r.Context().Value(middleware.UserIDCtxName); id != nil {
			userIDCtx = id.(string)
		}

		if userIDCtx == "" {
			http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		}

		links, err := h.st.GetUserLinks(userIDCtx)
		if err != nil || len(links) == 0 {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		responseURLs := make([]ResponseFullURL, 0, len(links))
		for short, orign := range links {
			responseURL := ResponseFullURL{
				OriginalURL: orign,
				ShortURL:    h.config.BaseURL + short,
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

//GetPing checks the connection to the database
func (h *URLHandler) GetPing() http.HandlerFunc {
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

//DeleteUserLinksBatch accepts a list of abbreviated URL IDs to delete
func (h *URLHandler) DeleteUserLinksBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userIDCtx string
		if id := r.Context().Value(middleware.UserIDCtxName); id != nil {
			userIDCtx = id.(string)
		}

		if userIDCtx == "" {
			http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		}

		deleteURLs := make([]string, 0)
		if err := json.NewDecoder(r.Body).Decode(&deleteURLs); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// perform asynchronous deletion
		h.st.Delete(deleteURLs, userIDCtx)
		w.WriteHeader(http.StatusAccepted)
	}
}
