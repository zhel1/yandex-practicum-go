package server

import (
	"github.com/zhel1/yandex-practicum-go/internal/handlers"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"log"
	"net/http"
)

type Server struct {
	Addr string
}

func (s *Server) StartServer() {
	dbConn := storage.NewDBConn()

	server := &http.Server{
		Addr:    s.Addr,
		Handler: handlers.NewRouter(dbConn, s.Addr),
	}

	log.Fatalln(server.ListenAndServe())
}
