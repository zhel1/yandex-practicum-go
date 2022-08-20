// Package app contains methods for launching shortener service.
package app

import (
	"github.com/zhel1/yandex-practicum-go/internal/http"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"github.com/zhel1/yandex-practicum-go/internal/storage/infile"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inmemory"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inpsql"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
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

	deps := service.Deps{
		Storage: strg,
		BaseURL: cfg.BaseURL,
	}

	services := service.NewServices(deps)
	crypto := utils.NewCrypto(cfg.UserKey)
	handlers := http.NewHandler(services, crypto)

	// HTTP Server
	server := server.NewServer(&cfg, handlers.Init())
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
