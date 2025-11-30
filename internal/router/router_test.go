package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPostHandler(t *testing.T) {
	ts := httptest.NewServer(Init())
	defer ts.Close()

	tests := []struct {
		name         string // description of this test case
		method       string
		url          string
		expectedCode int
	}{
		// TODO: Add test cases.
		{name: "postSuccs", method: http.MethodPost, url: "http://yandex.ru/",
			expectedCode: http.StatusCreated},

		{name: "delError", method: http.MethodDelete, url: "http://google.ru/",
			expectedCode: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		res, _ := testRequest(t, ts, tt.method, tt.url, ts.URL)

		assert.Equal(t, tt.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")
	}
}

func TestIDGetHandler(t *testing.T) {
	ts := httptest.NewServer(Init())
	defer ts.Close()

	url := "https://practicum.yandex.ru/"

	res, shortURL := testRequest(t, ts, http.MethodPost, url, ts.URL)
	assert.Equal(t, http.StatusCreated, res.StatusCode, "Код ответа не совпадает с ожидаемым")

	tests := []struct {
		name         string // description of this test case
		method       string
		url          string
		expectedCode int
	}{
		// TODO: Add test cases.
		{name: "getSuccs", method: http.MethodGet, url: shortURL,
			expectedCode: http.StatusTemporaryRedirect},

		{name: "getError", method: http.MethodGet, url: shortURL[:23] + "djhghhj",
			expectedCode: http.StatusNotFound},

		{name: "getError2", method: http.MethodGet, url: shortURL[:23],
			expectedCode: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		res, _ := testRequest(t, ts, tt.method, "", tt.url)
		loc := res.Header.Get("Location")

		assert.Equal(t, tt.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")
		if res.StatusCode == http.StatusTemporaryRedirect {
			assert.Equal(t, url, loc, "URL определен неверно")
		}
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, body, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, path, strings.NewReader(body))
	require.NoError(t, err)

	cl := ts.Client()

	cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// Если сделать редирект больше 1, то функция IDGetHandler начинает всегда возвращать статус 200.
		// Это связано с вызовом функции service.GetFullURL(ID). Побороть эту проблему не получилось. Причина такой работы тоже не ясна.
		// Нашел выход - установить редирект = 1 т.к. первый проход функции IDGetHandler всегда возвращается с корректным результатом.
		if len(via) >= 1 { // Лимит
			return http.ErrUseLastResponse // Вернуть последний ответ без дальнейшего следования
		}
		return nil // Разрешить следующий редирект
	}

	resp, err := cl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
