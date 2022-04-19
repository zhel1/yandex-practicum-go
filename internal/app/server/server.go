package server

import (
	"log"
	"net/http"
)

//var Database = make(map[string]string)

type Server struct {
	Addr     string
	Database map[string]string
}

func (s *Server) StartServer() {
	http.HandleFunc("/", s.CommonHandler)

	server := &http.Server{
		Addr: s.Addr,
	}

	s.Database = make(map[string]string)

	log.Fatalln(server.ListenAndServe())
}