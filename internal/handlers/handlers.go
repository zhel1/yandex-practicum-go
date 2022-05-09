package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"io"
	"log"
	"net/http"
	"net/url"
)

func NewRouter(st storage.Storage, baseURL string) chi.Router {
	r := chi.NewRouter()
	r.Post("/", AddLink(st, baseURL))
	r.Post("/api/shorten", AddLinkJSON(st, baseURL))
	r.Get("/{id}", GetLink(st, baseURL))
	return r
}

func AddLink(st storage.Storage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		longLinkBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		longLink := string(longLinkBytes)

		if _, err = url.ParseRequestURI(longLink); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortIDLink := utils.MD5(longLink)[:8]
		err = st.Put(shortIDLink, longLink)
		if err != nil {
			log.Panicln(err)
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, baseURL+ shortIDLink)
	}
}

func AddLinkJSON(st storage.Storage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		b := struct {
			URL string `json:"url"`
		}{
			URL: "",
		}

		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := url.ParseRequestURI(b.URL); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortIDLink := utils.MD5(b.URL)[:8]
		err := st.Put(shortIDLink, b.URL)
		if err != nil {
			log.Panicln(err)
		}

		res := struct {
			Result string `json:"result"`
		}{
			Result: baseURL + shortIDLink,
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
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Location", longLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}
