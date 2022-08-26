// Package app contains methods for launching shortener service.
package app

import (
	"github.com/zhel1/yandex-practicum-go/internal/auth"
	"github.com/zhel1/yandex-practicum-go/internal/http"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"github.com/zhel1/yandex-practicum-go/internal/storage/infile"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inmemory"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inpsql"
	"log"

	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/server"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
)

func Run() {
	var cfg config.Config
	err := cfg.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var strg storage.Storage
	if cfg.DatabaseDSN != "" {
		strg, err = inpsql.NewStorage(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Database is used")
	} else if cfg.FileStoragePath != "" {
		strg, err = infile.NewStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("File is used")
	} else {
		strg = inmemory.NewStorage()
		log.Println("Memory is used")
	}
	defer strg.Close()

	tokenManager, err := auth.NewManager(cfg.UserKey)
	if err != nil {
		log.Fatal(err)
	}

	deps := service.Deps{
		Storage:      strg,
		BaseURL:      cfg.BaseURL,
		TokenManager: tokenManager,
	}

	services := service.NewServices(deps)
	handlers := http.NewHandler(services)

	// HTTP Server
	server := server.NewServer(&cfg, handlers.Init())
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
