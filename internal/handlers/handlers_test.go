package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	//"io"
	"io/ioutil"
	//"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestGetLink(t *testing.T) {
	db := storage.NewDBConn()
	db.ShortURL["1234567"] = "https://yandex.ru/news/story/Minoborony_zayavilo_ob_unichtozhenii_podLvovom_sklada_inostrannogo_oruzhiya--5da2bb9cc9ddc47c0adb17be6d81bd72?lang=ru&rubric=index&fan=1&stid=yjizNz0bbyG1LTQtz2jv&t=1650312349&tt=true&persistent_id=192628644&story=4bc48b1b-a772-571f-a583-40d87f145dd6"
	db.ShortURL["1234568"] = "https://yandex.ru/news/"

	addr := "localhost:8080"
	r := NewRouter(db, addr)
	ts := httptest.NewServer(r)
	defer ts.Close()

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
			wantCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, http.MethodGet, "/"+tt.value, "")
			assert.Equal(t, tt.wantCode, resp.StatusCode)

			if resp.StatusCode == http.StatusTemporaryRedirect {
				_, err := url.ParseRequestURI(resp.Header.Get("Location"))
				require.NoError(t, err)
			}

		})
	}
}

func TestAddLink(t *testing.T) {
	db := storage.NewDBConn()
	addr := "localhost:8080"
	ts := httptest.NewServer(NewRouter(db, addr))
	defer ts.Close()

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
		t.Run(tt.name, func(t *testing.T) {
			db.ShortURL = make(map[string]string)
			resp, body := testRequest(t, ts, http.MethodPost, "/", tt.body)
			assert.Equal(t, tt.want.code, resp.StatusCode)

			if tt.want.code != http.StatusBadRequest {
				_, err := url.ParseRequestURI(body)
				require.NoError(t, err)

				if _, valid := db.ShortURL[strings.Replace(body, "http://"+addr+"/", "", -1)]; !valid {
					t.Errorf("Link %s was not saved in database", body)
				}

				assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}
