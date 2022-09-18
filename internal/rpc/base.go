package rpc

import (
	"context"
	"errors"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	pb "github.com/zhel1/yandex-practicum-go/internal/rpc/proto"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
)

// UsersServer поддерживает все необходимые методы сервера.
type ShortenerServer struct {
	// you need to embed the pb.Unimplemented<TypeName> type
	// for compatibility with future versions
	pb.UnimplementedShortenerServer
	services *service.Services
}

func NewBaseServer(services *service.Services) *ShortenerServer {
	return &ShortenerServer{
		services: services,
	}
}

func (s *ShortenerServer) AddLink(ctx context.Context, in *pb.AddLinkRequest) (*pb.AddLinkResponse, error) {
	userID, err := utils.ExtractValueFromContext(ctx, dto.UserIDCtxName)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	shortLink, err := s.services.Shorten.ShortenURL(ctx, userID, dto.ModelOriginalURL{
		OriginalURL: in.LongLink,
	})

	if err != nil {
		switch {
		case errors.Is(err, dto.ErrAlreadyExists):
			return &pb.AddLinkResponse{
				ShortLink: shortLink.ShortURL,
			}, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.AddLinkResponse{
		ShortLink: shortLink.ShortURL,
	}, nil
}

func (s *ShortenerServer) GetLink(ctx context.Context, in *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	u, err := url.Parse(in.ShortLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(u.Path) < 2 {
		return nil, status.Error(codes.InvalidArgument, "bad short id")
	}

	originalLink, err := s.services.Users.GetOriginalURLByShort(ctx, u.Path[1:])
	if err != nil {
		switch err {
		case dto.ErrDeleted:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.GetLinkResponse{
		LongLink: originalLink,
	}, nil
}

func (s *ShortenerServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.services.Users.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.PingResponse{}, nil
}

func (s *ShortenerServer) AddLinkBatch(ctx context.Context, in *pb.AddLinkBatchRequest) (*pb.AddLinkBatchResponse, error) {
	userID, err := utils.ExtractValueFromContext(ctx, dto.UserIDCtxName)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	bReq := make([]dto.ModelOriginalURLBatch, 0, len(in.Items))
	for _, item := range in.Items {
		bReq = append(bReq, dto.ModelOriginalURLBatch{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.LongUrl,
		})
	}

	bResArr, err := s.services.Shorten.ShortenBatchURL(ctx, userID, bReq)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	respItems := make([]*pb.AddLinkBatchResponse_Item, 0, len(bResArr))
	for _, item := range bResArr {
		respItems = append(respItems, &pb.AddLinkBatchResponse_Item{
			CorrelationId: item.CorrelationID,
			ShortUrl:      item.ShortURL,
		})
	}

	return &pb.AddLinkBatchResponse{
		Items: respItems,
	}, nil
}

func (s *ShortenerServer) GetUserLinks(ctx context.Context, in *pb.GetUsersLinksRequest) (*pb.GetUsersLinksResponse, error) {
	userID, err := utils.ExtractValueFromContext(ctx, dto.UserIDCtxName)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	modelURLs, err := s.services.Users.GetURLsByUserID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}

	if len(modelURLs) == 0 {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	respItems := make([]*pb.GetUsersLinksResponse_Item, 0, len(modelURLs))
	for _, item := range modelURLs {
		respItems = append(respItems, &pb.GetUsersLinksResponse_Item{
			LongUrl:  item.OriginalURL,
			ShortUrl: item.ShortURL,
		})
	}

	return &pb.GetUsersLinksResponse{
		Items: respItems,
	}, nil
}

func (s *ShortenerServer) DeleteUserLinksBatch(ctx context.Context, in *pb.DeleteUserLinksBatchRequest) (*pb.DeleteUserLinksBatchResponse, error) {
	userID, err := utils.ExtractValueFromContext(ctx, dto.UserIDCtxName)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	err = s.services.Users.DeleteBatchURL(ctx, userID, in.Items)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &pb.DeleteUserLinksBatchResponse{}, nil
}

func (s *ShortenerServer) GetStatistic(ctx context.Context, in *pb.GetStatRequest) (*pb.GetStatResponse, error) {
	stat, err := s.services.Internal.GetStatistic(ctx)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &pb.GetStatResponse{
		UrlsCont:  int32(stat.URLsCount),
		UsersCont: int32(stat.URLsCount),
	}, nil
}
