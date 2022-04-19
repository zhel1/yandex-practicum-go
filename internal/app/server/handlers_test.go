package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestServer_RestoreLinkHandler(t *testing.T) {
	s := &Server{
		Addr: "localhost:8080",
	}

	s.Database = make(map[string]string)

	s.Database["1234567"] = "https://yandex.ru/news/story/Minoborony_zayavilo_ob_unichtozhenii_podLvovom_sklada_inostrannogo_oruzhiya--5da2bb9cc9ddc47c0adb17be6d81bd72?lang=ru&rubric=index&fan=1&stid=yjizNz0bbyG1LTQtz2jv&t=1650312349&tt=true&persistent_id=192628644&story=4bc48b1b-a772-571f-a583-40d87f145dd6"
	s.Database["1234568"] = "https://yandex.ru/news/"

	type want struct {
		code        	int
	}

	tests := []struct {
		name string
		link string
		want want
	}{
		// определяем все тесты
		{
			name: "positive test #1",
			link: "/1234567",
			want: want{
				code:        http.StatusTemporaryRedirect,
			},
		},
		{
			name: "positive test #2",
			link: "/1234568",
			want: want{
				code:        http.StatusTemporaryRedirect,
			},
		},
		{
			name: "negative test #3",
			link: "http://" + s.Addr,
			want: want{
				code:        http.StatusBadRequest,
			},
		},
		{
			name: "negative test #4",
			link: "http://" + s.Addr,
			want: want{
				code:        http.StatusBadRequest,
			},
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.link, nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.RestoreLinkHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			if len(s.Database) == 0 {
				t.Errorf("Database is empty")
			}

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			if tt.want.code != http.StatusBadRequest {
				//проверка на корректность url
				longurl := res.Header.Get("Location")
				_, err := url.ParseRequestURI(longurl)
				if err != nil  {
					t.Errorf("Bad URL was in the Database: %s", longurl)
				}
			}
		})
	}
}

func TestServer_ShortLinkHandler(t *testing.T) {
	s := &Server{
		Addr: "localhost:8080",
	}

	type want struct {
		code        int
		contentType string
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
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "positive test #2",
			body: "https://yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "negative test #3",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "negative test #4",
			body: "12312343214",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.Database = make(map[string]string)

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.ShortLinkHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			if tt.want.code != http.StatusBadRequest {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				_, err = url.ParseRequestURI(string(resBody))
				if err != nil  {
					t.Errorf("Bad URL was made: %s", string(resBody))
				}

				if _, valid := s.Database[strings.Replace(string(resBody), "http://" + s.Addr + "/" ,"" ,-1)]; !valid {
					t.Errorf("Link %s was not saves in Database", string(resBody))
				}

				if res.Header.Get("Content-Type") != tt.want.contentType {
					t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
				}
			}
		})
	}
}