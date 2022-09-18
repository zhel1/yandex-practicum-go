package rpc

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zhel1/yandex-practicum-go/internal/auth"
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/rpc/interceptors"
	pb "github.com/zhel1/yandex-practicum-go/internal/rpc/proto"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/storage/inmemory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/url"
	"strings"
	"testing"
)

type HandlersTestSuite struct {
	suite.Suite
	storage  storage.Storage
	cfg      *config.Config
	ts       *grpc.Server
	baseURL  string
	services *service.Services
}

func (ht *HandlersTestSuite) SetupTest() {
	cfg := config.Config{}
	cfg.Addr = "localhost:8080"
	cfg.BaseURL = "http://localhost:8080/"
	cfg.FileStoragePath = ""
	cfg.UserKey = "PaSsW0rD"

	tokenManager, err := auth.NewManager(cfg.UserKey)
	if err != nil {
		log.Fatal(err)
	}

	deps := service.Deps{
		Storage:      inmemory.NewStorage(),
		BaseURL:      cfg.BaseURL,
		TokenManager: tokenManager,
	}

	services := service.NewServices(deps)

	// TODO Get addr from config
	listener, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}

	midl := interceptors.InitInterceptors(services)

	ht.cfg = &cfg
	ht.storage = deps.Storage
	ht.baseURL = "http://localhost:8080/"
	ht.services = services
	ht.ts = grpc.NewServer(grpc.UnaryInterceptor(midl.UserIDInterceptor))
	pb.RegisterShortenerServer(ht.ts, NewBaseServer(services))
	go ht.ts.Serve(listener)
}

func (ht *HandlersTestSuite) TearDownTest() {
	ht.ts.GracefulStop()
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (ht *HandlersTestSuite) TestAddLink() {
	// TODO Get addr from config
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewShortenerClient(conn)

	type want struct {
		code codes.Code
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test #1",
			body: "https://yandex.ru/news/story/Minoborony_zayavilo_ob_unichtozhenii_podLvovom_sklada_inostrannogo_oruzhiya--5da2bb9cc9ddc47c0adb17be6d81bd72?lang=ru&rubric=index&fan=1&stid=yjizNz0bbyG1LTQtz2jv&t=1650312349&tt=true&persistent_id=192628644&story=4bc48b1b-a772-571f-a583-40d87f145dd6",
			want: want{
				code: codes.OK,
			},
		},
		{
			name: "negative test #2",
			body: "",
			want: want{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "negative test #3",
			body: "12312343214",
			want: want{
				code: codes.InvalidArgument,
			},
		},
	}

	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			var trailer metadata.MD
			res, err := client.AddLink(context.Background(), &pb.AddLinkRequest{
				LongLink: tt.body,
			}, grpc.Trailer(&trailer))

			if tt.want.code == codes.OK {
				require.NoError(t, err)
				_, err := url.ParseRequestURI(res.ShortLink)
				require.NoError(t, err)

				//get only id and check in db
				id := strings.Replace(res.ShortLink, ht.cfg.BaseURL, "", -1)
				_, err = ht.storage.Get(context.Background(), id)
				require.NoError(t, err)

				//check token
				values := trailer.Get(dto.UserIDCtxName.String())
				require.NotEqual(t, 0, len(values))

				if len(values) > 0 {
					require.NotEqual(t, "", values[0])
				}
			} else {
				e, ok := status.FromError(err)
				assert.Equal(t, true, ok)
				if ok {
					assert.Equal(t, tt.want.code, e.Code())
				}
			}
		})
	}
}

func (ht *HandlersTestSuite) TestGetLink() {
	// TODO Get addr from config
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewShortenerClient(conn)

	userID := uuid.New().String()
	ht.storage.Put(context.Background(), userID, "1234567", "https://yandex.ru/news/story/Minoborony_zayavilo_ob_unichtozhenii_podLvovom_sklada_inostrannogo_oruzhiya--5da2bb9cc9ddc47c0adb17be6d81bd72?lang=ru&rubric=index&fan=1&stid=yjizNz0bbyG1LTQtz2jv&t=1650312349&tt=true&persistent_id=192628644&story=4bc48b1b-a772-571f-a583-40d87f145dd6")
	ht.storage.Put(context.Background(), userID, "1234568", "https://yandex.ru/news/")

	tests := []struct {
		name     string
		value    string
		wantCode codes.Code
	}{
		{
			name:     "Positive test #1",
			value:    "1234567",
			wantCode: codes.OK,
		},
		{
			name:     "Positive test #2",
			value:    "1234568",
			wantCode: codes.OK,
		},
		{
			name:     "Negative test #3. No link in database.",
			value:    "1234569",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Negative test #4 . Not existing path.",
			value:    "1234567/1234567",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Negative test #5. Empty path (redirection test).",
			value:    "",
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			res, err := client.GetLink(context.Background(), &pb.GetLinkRequest{
				ShortLink: ht.baseURL + tt.value,
			})

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				_, err := url.ParseRequestURI(res.LongLink)
				require.NoError(t, err)
			} else {
				e, ok := status.FromError(err)
				assert.Equal(t, true, ok)
				if ok {
					assert.Equal(t, tt.wantCode, e.Code())
				}
			}
		})
	}
}

func (ht *HandlersTestSuite) TestPing() {
	// TODO Get addr from config
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewShortenerClient(conn)

	ht.T().Run("Ping", func(t *testing.T) {
		_, err := client.Ping(context.Background(), &pb.PingRequest{})
		require.NoError(t, err)
	})
}

func (ht *HandlersTestSuite) TestAddLinkBatch() {
	// TODO Get addr from config
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewShortenerClient(conn)

	type want struct {
		code codes.Code
	}

	tests := []struct {
		name string
		body *pb.AddLinkBatchRequest
		want want
	}{
		{
			name: "positive test #1",
			body: &pb.AddLinkBatchRequest{
				Items: []*pb.AddLinkBatchRequest_Item{
					{
						CorrelationId: "1",
						LongUrl:       "http://khawesxujm.biz/fdapyknrhiywl0",
					},
					{
						CorrelationId: "2",
						LongUrl:       "http://jlth8bcthyp01q.ru/zkd2d",
					},
				},
			},
			want: want{
				code: codes.OK,
			},
		},
		{
			name: "negative test #2",
			body: &pb.AddLinkBatchRequest{},
			want: want{
				code: codes.InvalidArgument,
			},
		},
	}

	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			var trailer metadata.MD
			res, err := client.AddLinkBatch(context.Background(), tt.body, grpc.Trailer(&trailer))
			if tt.want.code == codes.OK {
				require.NoError(t, err)
				for _, item := range res.Items {
					u, err := url.ParseRequestURI(item.ShortUrl)
					require.NoError(t, err)

					require.NotEqual(t, 0, len(u.Path))
					require.NotEqual(t, 1, len(u.Path))

					_, err = ht.storage.Get(context.Background(), u.Path[1:])
					require.NoError(t, err)
				}

				//check token
				values := trailer.Get(dto.UserIDCtxName.String())
				require.NotEqual(t, 0, len(values))

				if len(values) > 0 {
					require.NotEqual(t, "", values[0])
				}
			} else {
				e, ok := status.FromError(err)
				assert.Equal(t, true, ok)
				if ok {
					assert.Equal(t, tt.want.code, e.Code())
				}
			}
		})
	}
}

func (ht *HandlersTestSuite) TestGetUserLinks() {
	// TODO Get addr from config
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewShortenerClient(conn)

	userID := uuid.New().String()
	token, err := ht.services.Users.CreateNewToken(context.Background(), userID)
	if err != nil {
		panic(err.Error())
	}

	idLink := "1234568"
	origLink := "https://yandex.ru/news/"

	tests := []struct {
		name     string
		value    string
		stData   bool
		wantCode codes.Code
	}{
		{
			name:     "Negative test #1. No user in database.",
			stData:   false,
			wantCode: codes.NotFound,
		},
		{
			name:     "Positive test #2",
			stData:   true,
			wantCode: codes.OK,
		},
	}

	for _, tt := range tests {
		if tt.stData {
			ht.storage.Put(context.Background(), userID, idLink, origLink)
		}
		ht.T().Run(tt.name, func(t *testing.T) {
			md := metadata.New(map[string]string{dto.UserIDCtxName.String(): token})
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			res, err := client.GetUserLinks(ctx, &pb.GetUsersLinksRequest{})
			if tt.stData {
				require.NoError(t, err)
				require.Equal(t, origLink, res.Items[0].LongUrl)

				u, err := url.ParseRequestURI(res.Items[0].ShortUrl)
				require.NoError(t, err)
				require.Equal(t, "/"+idLink, u.Path)
			} else {
				e, ok := status.FromError(err)
				assert.Equal(t, true, ok)
				if ok {
					assert.Equal(t, tt.wantCode, e.Code())
				}
			}
		})
	}
}
