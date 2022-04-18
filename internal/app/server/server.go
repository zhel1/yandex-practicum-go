package server

import (
	"log"
	"net/http"
)

var dataBase = make(map[string]string)

type Server struct {
	Addr string
}

func (s *Server) StartServer() {
	http.HandleFunc("/", s.CommonHandler)

	server := &http.Server{
		Addr: s.Addr,
	}

	log.Fatalln(server.ListenAndServe())
}