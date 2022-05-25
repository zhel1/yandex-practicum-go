package server

import (
	"github.com/zhel1/yandex-practicum-go/internal/handlers"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"log"
	"net/http"
)

type Server struct {
	Addr    string
	BaseURL string
	Storage storage.Storage
}

func (s *Server) StartServer() {
	server := &http.Server {
		Addr:    s.Addr,
		Handler: handlers.NewRouter(s.Storage, s.BaseURL),
	}

	log.Fatalln(server.ListenAndServe())
}
