package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/handlers"
	"github.com/zhel1/yandex-practicum-go/internal/middleware"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"log"
	"net/http"
)

type Server struct {
	Config *config.Config
	Storage storage.Storage
}

func (s *Server) StartServer() error {
	URLHandler, err := handlers.InitURLHandler(s.Storage, s.Config)
	if err != nil {
		return err
	}

	crypto, err := utils.NewCrypto(s.Config.UserKey)
	if err != nil {
		return err
	}

	cookieHandler, err := middleware.NewCookieHandler(crypto)
	if err != nil {
		return err
	}

	r := chi.NewRouter()
	r.Use(middleware.GzipHandle)
	r.Use(cookieHandler.CokieHandle)
	r.Post("/", URLHandler.AddLink())
	r.Post("/api/shorten", URLHandler.AddLinkJSON())
	r.Post("/api/batch", URLHandler.AddLinkBatchJSON())
	r.Get("/api/user/urls", URLHandler.GetUserLinks())
	r.Get("/{id}", URLHandler.GetLink())
	r.Get("/ping", URLHandler.GetPing())

	server := &http.Server {
		Addr:    s.Config.Addr,
		Handler: r,
	}

	log.Fatalln(server.ListenAndServe())
	return nil
}
