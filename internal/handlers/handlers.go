package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"io"
	"log"
	"net/http"
	"net/url"
)

func NewRouter(db *storage.DB, addr string) chi.Router {
	r := chi.NewRouter()
	r.Post("/", AddLink(db, addr))
	r.Get("/{id}", GetLink(db, addr))
	return r
}

func AddLink(db *storage.DB, addr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		longLinkBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		longLink := string(longLinkBytes)

		_, err = url.ParseRequestURI(longLink)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortIDLink := utils.MD5(longLink)[:8]
		_, valid := db.ShortURL[shortIDLink]
		if !valid {
			db.ShortURL[shortIDLink] = longLink
		} else {
			log.Panicln("Ссылка существует")
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "http://"+addr+"/"+shortIDLink)
	}

}

func GetLink(db *storage.DB, addr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		linkId := chi.URLParam(r, "id")
		longLink, valid := db.ShortURL[linkId]
		if valid {
			w.Header().Set("Location", longLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
