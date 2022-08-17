package main

import (
	"github.com/zhel1/yandex-practicum-go/internal/app/http"
	"github.com/zhel1/yandex-practicum-go/internal/app/service"
	"github.com/zhel1/yandex-practicum-go/internal/app/storage/infile"
	"github.com/zhel1/yandex-practicum-go/internal/app/storage/inmemory"
	"github.com/zhel1/yandex-practicum-go/internal/app/storage/inpsql"
	"github.com/zhel1/yandex-practicum-go/internal/app/utils"
	"log"

	"github.com/zhel1/yandex-practicum-go/internal/app/config"
	"github.com/zhel1/yandex-practicum-go/internal/app/server"
	"github.com/zhel1/yandex-practicum-go/internal/app/storage"
)

func main() {
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
