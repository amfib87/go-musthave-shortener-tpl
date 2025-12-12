package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMainPostHandler(t *testing.T) {
	tests := []struct {
		name         string // description of this test case
		cfg          *config.Cnfg
		method       string
		url          string
		expectedCode int
	}{
		// TODO: Add test cases.
		{name: "postSuccs", cfg: &config.Cnfg{ServRunAddr: "", AddrForURL: ""}, method: http.MethodPost,
			url: "http://yandex", expectedCode: http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.cfg)
			res := httptest.NewRecorder()

			body := tt.url
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(body))

			h.MainPostHandler(res, req)

			assert.Equal(t, tt.expectedCode, res.Code, "код ответа не совпадает с ожидаемым")
		})
	}
}

func TestIDGetHandler(t *testing.T) {
	url := "http://rambler"
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))

	cfg := &config.Cnfg{ServRunAddr: "", AddrForURL: ""}
	h := NewHandler(cfg)

	h.MainPostHandler(res, req)
	assert.Equal(t, http.StatusCreated, res.Code, "код ответа не совпадает с ожидаемым")

	shortURL := res.Body.String()

	tests := []struct {
		name         string // description of this test case
		method       string
		url          string
		expectedCode int
	}{
		// TODO: Add test cases.
		{name: "getSuccs", method: http.MethodGet, url: shortURL,
			expectedCode: http.StatusTemporaryRedirect},

		{name: "getError", method: http.MethodGet, url: "http://rbc/hhtht",
			expectedCode: http.StatusInternalServerError},

		{name: "getError2", method: http.MethodGet, url: "http://1",
			expectedCode: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.url, nil)

			h.IDGetHandler(res, req)
			loc := res.Header().Get("Location")

			assert.Equal(t, tt.expectedCode, res.Code, "код ответа не совпадает с ожидаемым")
			if res.Code == http.StatusTemporaryRedirect {
				assert.Equal(t, url, loc, "url определен неверно")
			}
		})
	}
}
