package http

import (
	"encoding/json"
	"github.com/zhel1/yandex-practicum-go/internal/app/dto"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

const AddLinkEndpoint = "http://localhost:8080/"
const GetLinkEndpoint = "http://localhost:8080/" // + id

func Benchmark_AddLink(b *testing.B) {
	var URL string
	encryptedUserID := "03042d3702a2c3fac75929965e5ddb80775076c1cbcc4a456d19bf576abc24ff49ff2592f8f45d4540c38d6a9e1f05ca02dc7b93"
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  dto.UserIDCtxName.String(),
		Value: encryptedUserID,
		Path:  "/",
	})

	b.ResetTimer()
	b.Run("b", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := uuid.New().String()
			URL = "https://www." + id + ".com"

			reqBody, _ := json.Marshal(URL)
			payload := strings.NewReader(string(reqBody))

			b.StartTimer()
			_, _ = client.R().SetBody(payload).Post(AddLinkEndpoint)
		}
	})
}

func Benchmark_GetLink(b *testing.B) {
	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}))

	b.ResetTimer()
	b.Run("b", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			ShortURL := uuid.New().String()
			b.StartTimer()
			_, _ = client.R().Get(GetLinkEndpoint + ShortURL)
		}
	})
}
