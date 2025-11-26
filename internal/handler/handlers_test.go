package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainPostHandler(t *testing.T) {
	tests := []struct {
		name         string // description of this test case
		method       string
		url          string
		expectedCode int
	}{
		// TODO: Add test cases.
		{name: "getError", method: http.MethodGet, url: "http://rambler",
			expectedCode: http.StatusMethodNotAllowed},

		{name: "postSuccs", method: http.MethodPost, url: "http://yandex",
			expectedCode: http.StatusCreated},

		{name: "delError", method: http.MethodDelete, url: "http://google",
			expectedCode: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			body := tt.url
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(body))

			MainPostHandler(res, req)

			assert.Equal(t, tt.expectedCode, res.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestIDGetHandler(t *testing.T) {
	url := "http://rambler"
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))

	MainPostHandler(res, req)
	assert.Equal(t, http.StatusCreated, res.Code, "Код ответа не совпадает с ожидаемым")

	shortURL := res.Body.String()

	tests := []struct {
		name         string // description of this test case
		method       string
		url          string
		expectedCode int
	}{
		// TODO: Add test cases.
		{name: "postError", method: http.MethodPost, url: "http://error",
			expectedCode: http.StatusMethodNotAllowed},

		{name: "getSuccs", method: http.MethodGet, url: shortURL,
			expectedCode: http.StatusTemporaryRedirect},

		{name: "getError", method: http.MethodGet, url: "http://rbc/hhtht",
			expectedCode: http.StatusNotFound},

		{name: "getError2", method: http.MethodGet, url: "http://1",
			expectedCode: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.url, nil)

			IDGetHandler(res, req)
			loc := res.Header().Get("Location")

			assert.Equal(t, tt.expectedCode, res.Code, "Код ответа не совпадает с ожидаемым")
			if res.Code == http.StatusTemporaryRedirect {
				assert.Equal(t, url, loc, "URL определен неверно")
			}
		})
	}
}
