package main

import (
	"flag"
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

func main() {
	var cfg Config
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "host to listen on")
	flag.StringVar(&cfg.BaseURL,"b", "http://localhost:8080", "Base address of the resulting shortened URL")
	flag.StringVar(&cfg.FileStoragePath,"f", "", "Path to the file with shortened URLs")
	flag.Parse()

	//settings redefinition, if evn variables is used
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
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
		BaseURL: cfg.BaseURL + "/",
		Storage: strg,
	}
	s.StartServer()
}
