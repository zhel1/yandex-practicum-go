package main

import (
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/server"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"log"
)

func main() {
	var cfg config.Config
	err := cfg.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var strg storage.Storage
	if cfg.DatabaseDSN != "" {
		strg, err = storage.NewInPSQL(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
	} else if cfg.FileStoragePath != "" {
		strg, err = storage.NewInFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		strg = storage.NewInMemory()
	}
	defer strg.Close()

	s := server.Server {
		Config: &cfg,
		Storage: strg,
	}
	err = s.StartServer()
	if err != nil {
		log.Fatal(err)
	}
}


