package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/zhel1/yandex-practicum-go/internal/server"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"log"
)

type Config struct {
	Addr			string		`env:"SERVER_ADDRESS"`
	BaseURL			string		`env:"BASE_URL"`
	FileStoragePath	string		`env:"FILE_STORAGE_PATH"`
}

func NewConfig() Config {
	return Config {
		Addr: "localhost:8080",
		BaseURL: "http://localhost:8080/",
		FileStoragePath: "",
	}
}

func main() {
	cfg := NewConfig()
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	if cfg.Addr == "" || cfg.BaseURL == "" {
		log.Fatal("BASE_URL or SERVER_ADDRESS env variables not found.")
	}

	var strg storage.Storage
	if cfg.FileStoragePath != "" {
		strg, err = storage.NewInFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		strg = storage.NewInMemory()
	}
	defer strg.Close()

	s := server.Server {
		Addr:    cfg.Addr,
		BaseURL: cfg.BaseURL,
		Storage: strg,
	}
	s.StartServer()
}
