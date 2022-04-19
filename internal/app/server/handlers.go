package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func (s *Server)CommonHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// если методом POST
	case http.MethodPost:
		s.ShortLinkHandler(w, r)
	case http.MethodGet:
		s.RestoreLinkHandler(w, r)
	default:
		// RESPONSE
		w.Header().Set("content-type","text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w,  "Error: Method not allowed")
	}
}

// Эндпоинт POST / принимает в теле запроса строку URL для
// сокращения и возвращает ответ с кодом 201 и сокращённым
// URL в виде текстовой строки в теле.
func (s *Server)ShortLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	// читаем Body
	longLinkBytes, err := io.ReadAll(r.Body)
	longLink := string(longLinkBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//проверка на корректность url
	_, err = url.ParseRequestURI(longLink)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Сокращаем и сохраняем
	shortIDLink := MD5(string(longLink))[:8]
	_, valid := s.Database[shortIDLink]
	if !valid {
		s.Database[shortIDLink] = string(longLink)
	} else {
		log.Panicln("Ссылка существует")
	}

	// RESPONSE
	w.Header().Set("content-type","text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w,  "http://" + s.Addr + "/" +shortIDLink)
}

// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор
// сокращённого URL и возвращает ответ с кодом 307 и
// оригинальным URL в HTTP-заголовке Location.
func (s *Server)RestoreLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	url := r.URL.RequestURI()

	if url != ""  {
		shortIDLink := url[1:]

		longLink, valid := s.Database[shortIDLink]
		if valid {
			// RESPONSE
			w.Header().Set("Location", longLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		// RESPONSE
		w.WriteHeader(http.StatusBadRequest)
	}
}