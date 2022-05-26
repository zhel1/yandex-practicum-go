package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"github.com/zhel1/yandex-practicum-go/internal/middleware"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"net/url"
	"strings"

	//"io"
	//"io/ioutil"
	//"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HandlersTestSuite struct {
	suite.Suite
	storage          storage.Storage
	cfg              *config.Config
	urlHandler       *URLHandler
	cookieHandler    *middleware.CookieHandler
	cookieEncriptor  *utils.Crypto
	router           *chi.Mux
	ts               *httptest.Server
}

func (ht *HandlersTestSuite) SetupTest() {
	cfg := config.Config{}
	cfg.Addr = "localhost:8080"
	cfg.BaseURL = "http://localhost:8080/"
	cfg.FileStoragePath = ""
	cfg.UserKey = "PaSsW0rD"

	ht.cfg = &cfg
	ht.cookieEncriptor, _ = utils.NewCrypto(cfg.UserKey)
	ht.storage = storage.NewInMemory()
	ht.urlHandler, _ = InitURLHandler(ht.storage, &cfg)
	ht.cookieHandler, _ = middleware.NewCookieHandler(ht.cookieEncriptor)
	ht.router = chi.NewRouter()
	ht.ts = httptest.NewServer(ht.router)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (ht *HandlersTestSuite)TestGetLink() {
	ht.router.Get("/{id}", ht.urlHandler.GetLink())
	userID := uuid.New().String()
	ht.storage.Put(userID,"1234567", "https://yandex.ru/news/story/Minoborony_zayavilo_ob_unichtozhenii_podLvovom_sklada_inostrannogo_oruzhiya--5da2bb9cc9ddc47c0adb17be6d81bd72?lang=ru&rubric=index&fan=1&stid=yjizNz0bbyG1LTQtz2jv&t=1650312349&tt=true&persistent_id=192628644&story=4bc48b1b-a772-571f-a583-40d87f145dd6")
	ht.storage.Put(userID,"1234568", "https://yandex.ru/news/")
	defer ht.ts.Close()

	tests := []struct {
		name     string
		value    string
		wantCode int
	}{
		{
			name:     "Positive test #1",
			value:    "1234567",
			wantCode: http.StatusTemporaryRedirect,
		},
		{
			name:     "Positive test #1",
			value:    "1234568",
			wantCode: http.StatusTemporaryRedirect,
		},
		{
			name:     "Negative test #2. No link in database.",
			value:    "1234569",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Negative test #3 . Not existing path.",
			value:    "1234567/1234567",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative test #4. Empty path (redirection test).",
			value:    "",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}))
			resp, err := client.R().Get(ht.ts.URL + "/" + tt.value)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, resp.StatusCode())

			if resp.StatusCode() == http.StatusTemporaryRedirect {
				_, err := url.ParseRequestURI(resp.RawResponse.Header.Get("Location"))
				require.NoError(t, err)
			}
		})
	}
}

func (ht *HandlersTestSuite)TestAddLink() {
	ht.router.Use(ht.cookieHandler.CokieHandle)
	ht.router.Post("/", ht.urlHandler.AddLink())
	defer ht.ts.Close()

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name string
		body string
		want want
	} {
		{
			name: "positive test #1",
			body: "https://yandex.ru/news/story/Minoborony_zayavilo_ob_unichtozhenii_podLvovom_sklada_inostrannogo_oruzhiya--5da2bb9cc9ddc47c0adb17be6d81bd72?lang=ru&rubric=index&fan=1&stid=yjizNz0bbyG1LTQtz2jv&t=1650312349&tt=true&persistent_id=192628644&story=4bc48b1b-a772-571f-a583-40d87f145dd6",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "negative test #2",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "negative test #3",
			body: "12312343214",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			resp, err := client.R().SetBody(tt.body).Post(ht.ts.URL + "/")
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode())

			if tt.want.code != http.StatusBadRequest {
				_, err := url.ParseRequestURI(string(resp.Body()))
				require.NoError(t, err)

				//get only id
				id := strings.Replace(string(resp.Body()), ht.cfg.BaseURL, "", -1)
				_, err = ht.storage.Get(id)
				require.NoError(t, err)

				assert.Equal(t, tt.want.contentType, resp.RawResponse.Header.Get("Content-Type"))
			}
		})
	}
}

func (ht *HandlersTestSuite)TestAddLinkJSON() {
	ht.router.Use(ht.cookieHandler.CokieHandle)
	ht.router.Post("/api/shorten", ht.urlHandler.AddLinkJSON())
	defer ht.ts.Close()

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name string
		body string
		want want
	} {
		{
			name: "positive test #1",
			body: "{\"url\":\"https://yandex.ru/news/story/Minoborony_zayavilo_ob_unichtozhenii_podLvovom_sklada_inostrannogo_oruzhiya--5da2bb9cc9ddc47c0adb17be6d81bd72?lang=ru&rubric=index&fan=1&stid=yjizNz0bbyG1LTQtz2jv&t=1650312349&tt=true&persistent_id=192628644&story=4bc48b1b-a772-571f-a583-40d87f145dd6\"}",
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "negative test #2",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "negative test #3",
			body: "12312343214",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			resp, err := client.R().SetBody(tt.body).Post(ht.ts.URL + "/api/shorten")
			require.NoError(t, err)
			//defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode())

			if tt.want.code != http.StatusBadRequest {
				res := struct {
					Result string `json:"result"`
				}{
					Result: "",
				}

				err := json.Unmarshal(resp.Body(), &res)
				require.NoError(t, err)

				_, err = url.ParseRequestURI(res.Result)
				require.NoError(t, err)

				//get only id
				id := strings.Replace(res.Result, ht.cfg.BaseURL, "", -1)
				_, err = ht.storage.Get(id)
				require.NoError(t, err)

				assert.Equal(t, tt.want.contentType, resp.RawResponse.Header.Get("Content-Type"))
			}
		})
	}
}

func (ht *HandlersTestSuite)TestGetUserLinks() {
	ht.router.Use(ht.cookieHandler.CokieHandle)
	ht.router.Get("/api/user/urls", ht.urlHandler.GetUserLinks())

	crypto, _ := utils.NewCrypto(ht.cfg.UserKey)
	userID := crypto.Encode(uuid.New().String())
	idLink := "1234568"
	origLink := "https://yandex.ru/news/"

	defer ht.ts.Close()

	tests := []struct {
		name     string
		value    string
		stData 	 bool
		wantCode int
	}{
		{
			name:     "Negative test #1. No user in database.",
			stData:  false,
			wantCode: http.StatusNoContent,
		},
		{
			name:     "Positive test #2",
			stData: true,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		if tt.stData { ht.storage.Put(userID, idLink, origLink) }
		ht.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetCookie(&http.Cookie{
				Name: middleware.UserIDCtxNameText(middleware.UserIDCtxName),
				Value: crypto.Encode(userID),
				Path:  "/",
			})
			resp, err := client.R().Get(ht.ts.URL + "/api/user/urls")
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, resp.StatusCode())
		})
	}
}