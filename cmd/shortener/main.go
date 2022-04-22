package main

import "github.com/zhel1/yandex-practicum-go/internal/server"

func main() {
	s := server.Server{
		Addr: "localhost:8080",
	}
	s.StartServer()
}
