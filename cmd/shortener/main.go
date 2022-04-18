package main

import "github.com/zhel1/yandex-practicum-go/internal/app/server"

func main() {

	serverAddr := "localhost:8080"

	s := server.Server {Addr: serverAddr}
	s.StartServer()
}