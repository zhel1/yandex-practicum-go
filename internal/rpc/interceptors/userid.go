package interceptors

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptors struct {
	services *service.Services
}

func InitInterceptors(services *service.Services) *Interceptors {
	if services == nil {
		panic(fmt.Errorf("nil services was passed to grpc interceptors"))
	}

	return &Interceptors{
		services: services,
	}
}

func (i *Interceptors) UserIDInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var userID string
	var token string
	var err error
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(dto.UserIDCtxName.String())
		if len(values) > 0 {
			userID, err = i.services.Users.CheckToken(ctx, values[0])
			if err == nil {
				return handler(context.WithValue(ctx, dto.UserIDCtxName, userID), req)
			}
		}
	}

	userID = uuid.New().String()
	token, err = i.services.Users.CreateNewToken(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `create token error:`+err.Error())
	}

	md := metadata.New(map[string]string{dto.UserIDCtxName.String(): token})
	err = grpc.SetTrailer(ctx, md)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `set trailer error:`+err.Error())
	}

	return handler(context.WithValue(ctx, dto.UserIDCtxName, userID), req)
}
