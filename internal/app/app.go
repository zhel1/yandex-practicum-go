// Package app contains methods for launching shortener service.
package app

import (
	"context"
	"github.com/zhel1/yandex-practicum-go/internal/auth"
	"github.com/zhel1/yandex-practicum-go/internal/http"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"github.com/zhel1/yandex-practicum-go/internal/storage/infile"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inmemory"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inpsql"
	"log"
	"net"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"

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

	tokenManager, err := auth.NewManager(cfg.UserKey)
	if err != nil {
		log.Fatal(err)
	}

	var ipnet *net.IPNet = nil
	if cfg.TrustedSubnet != "" {
		_, ipnet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			log.Fatal(err)
		}
	}

	deps := service.Deps{
		Storage:       strg,
		BaseURL:       cfg.BaseURL,
		TokenManager:  tokenManager,
		TrustedSubnet: ipnet,
	}

	services := service.NewServices(deps)
	handlers := http.NewHandler(services)

	// HTTP Server
	srv := server.NewServer(&cfg, handlers.Init())
	connectionsClosed := make(chan struct{})
	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-interrupt
		if err := srv.Stop(context.Background()); err != nil {
			log.Printf("HTTP server shutdown: %v", err)
		}

		if err := strg.Close(); err != nil {
			log.Printf("Storage shutdown: %v", err)
		}

		close(connectionsClosed)
	}()

	if err := srv.Run(); err != nethttp.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-connectionsClosed
	log.Println("Server shutdown gracefully")
}
