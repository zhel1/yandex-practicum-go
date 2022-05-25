package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func NewRouter(st storage.Storage, baseURL string) chi.Router {
	r := chi.NewRouter()
	r.Use(gzipHandle)
	r.Post("/", AddLink(st, baseURL))
	r.Post("/api/shorten", AddLinkJSON(st, baseURL))
	r.Get("/{id}", GetLink(st, baseURL))
	return r
}

func shortenURL(st storage.Storage, baseURL, URL string) (string, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return "", err
	}

	shortIDLink := utils.MD5(URL)[:8]

	if err := st.Put(shortIDLink, URL); err != nil {
		return "", err
	}
	return baseURL + shortIDLink, nil
}
//**********************************************************************************************************************
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get(`Content-Encoding`), `gzip`) {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gzReader
			defer gzReader.Close()
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzWriter, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gzWriter.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzWriter}, r)
	})
}
//**********************************************************************************************************************
func AddLink(st storage.Storage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		longLinkBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortLink, err := shortenURL(st, baseURL, string(longLinkBytes))
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

func AddLinkJSON(st storage.Storage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		b := JSONRequestData {}
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shortLink, err := shortenURL(st, baseURL, b.URL)
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
		encoder.Encode(res)

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, buf)
	}
}

func GetLink(st storage.Storage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		linkID := chi.URLParam(r, "id")
		longLink, err := st.Get(linkID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			w.Header().Set("Location", longLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}
