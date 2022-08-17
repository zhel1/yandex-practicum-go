package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/zhel1/yandex-practicum-go/internal/middleware"
)

const AddLinkEndpoint = "http://localhost:8080/"
const AddLinkJSONEndpoint = "http://localhost:8080/api/shorten"
const AddLinkBatchJSONEndpoint = "http://localhost:8080/api/shorten/batch"
const GetLinkEndpoint = "http://localhost:8080/" // + id
const GetUserLinksEndpoint = "http://localhost:8080/api/user/urls"
const DeleteUserLinksBatchEndpoint = "http://localhost:8080/api/user/urls"

func Benchmark_AddLink(b *testing.B) {

	var URL string
	encryptedUserID := "03042d3702a2c3fac75929965e5ddb80775076c1cbcc4a456d19bf576abc24ff49ff2592f8f45d4540c38d6a9e1f05ca02dc7b93"
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserIDCtxName.String(),
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

func Benchmark_AddLinkJSON(b *testing.B) {

	URL := JSONRequestData{}
	encryptedUserID := "03042d3702a2c3fac75929965e5ddb80775076c1cbcc4a456d19bf576abc24ff49ff2592f8f45d4540c38d6a9e1f05ca02dc7b93"
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserIDCtxName.String(),
		Value: encryptedUserID,
		Path:  "/",
	})

	b.ResetTimer()
	b.Run("b", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := uuid.New().String()
			URL = JSONRequestData{URL: "https://www." + id + ".com"}

			reqBody, _ := json.Marshal(URL)
			payload := strings.NewReader(string(reqBody))

			b.StartTimer()
			_, _ = client.R().SetBody(payload).Post(AddLinkJSONEndpoint)
		}
	})
}

func Benchmark_AddLinkBatchJSON(b *testing.B) {

	URLs := make([]BatchRequest, 20000)
	encryptedUserID := "03042d3702a2c3fac75929965e5ddb80775076c1cbcc4a456d19bf576abc24ff49ff2592f8f45d4540c38d6a9e1f05ca02dc7b93"
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserIDCtxName.String(),
		Value: encryptedUserID,
		Path:  "/",
	})

	b.ResetTimer()
	b.Run("b", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			URLs = nil
			for j := 0; j < 5000; j++ {
				id := uuid.New().String()
				URLs = append(URLs, BatchRequest{OriginalURL: "https://www." + id + ".com", CorrelationID: "user"})
			}
			reqBody, _ := json.Marshal(URLs)
			payload := strings.NewReader(string(reqBody))

			b.StartTimer()
			_, _ = client.R().SetBody(payload).Post(AddLinkBatchJSONEndpoint)
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

func Benchmark_GetUserLinks(b *testing.B) {
	encryptedUserID := "03042d3702a2c3fac75929965e5ddb80775076c1cbcc4a456d19bf576abc24ff49ff2592f8f45d4540c38d6a9e1f05ca02dc7b93"
	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}))
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserIDCtxName.String(),
		Value: encryptedUserID,
		Path:  "/",
	})
	b.ResetTimer()
	b.Run("b", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			b.StartTimer()
			_, _ = client.R().Get(GetUserLinksEndpoint)
		}
	})
}

func Benchmark_DeleteUserLinksBatch(b *testing.B) {

	ShortURLs := make([]string, 20000)
	encryptedUserID := "03042d3702a2c3fac75929965e5ddb80775076c1cbcc4a456d19bf576abc24ff49ff2592f8f45d4540c38d6a9e1f05ca02dc7b93"
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  middleware.UserIDCtxName.String(),
		Value: encryptedUserID,
		Path:  "/",
	})

	b.ResetTimer()
	b.Run("b", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			ShortURLs = nil
			for j := 0; j < 5000; j++ {
				id := uuid.New().String()
				ShortURLs = append(ShortURLs, id)
			}
			reqBody, _ := json.Marshal(ShortURLs)
			payload := strings.NewReader(string(reqBody))

			b.StartTimer()
			_, _ = client.R().SetBody(payload).Delete(DeleteUserLinksBatchEndpoint)
		}
	})
}
