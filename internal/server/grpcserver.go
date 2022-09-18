package server

import (
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/rpc"
	"github.com/zhel1/yandex-practicum-go/internal/rpc/interceptors"
	pb "github.com/zhel1/yandex-practicum-go/internal/rpc/proto"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

// GRPCServer struct
type GRPCServer struct {
	grpcServer  *grpc.Server
	listener    net.Listener
	enableHTTPS bool
}

func NewGRPCServer(cfg *config.Config, services *service.Services) *GRPCServer {
	server := &GRPCServer{}
	var err error

	//TODO Add addr in config
	server.listener, err = net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}

	i := interceptors.InitInterceptors(services)

	server.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(i.UserIDInterceptor))
	pb.RegisterShortenerServer(server.grpcServer, rpc.NewBaseServer(services))

	return server
}

func (s *GRPCServer) Run() error {
	return s.grpcServer.Serve(s.listener)
}

func (s *GRPCServer) Stop() {
	s.grpcServer.GracefulStop()
}
