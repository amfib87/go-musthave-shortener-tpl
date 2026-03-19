package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/audit"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPostHandler(t *testing.T) {
	tempFile, err := os.CreateTemp(os.TempDir(), "Iter9")
	if err != nil {
		t.Fatal("failed create temp file", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name()) // удаляем файл после теста
	}()
	path := tempFile.Name()

	tests := []struct {
		name         string // description of this test case
		cfg          *config.Cnfg
		method       string
		url          string
		expectedCode int
	}{
		{name: "postSuccs", cfg: &config.Cnfg{ServRunAddr: "", AddrForURL: "", StoragePath: path}, method: http.MethodPost,
			url: "http://yandex", expectedCode: http.StatusCreated},
	}

	logger, err := logger.Initialize("Info")
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewHandler(tt.cfg, logger, service.URLStorage{}, &audit.AuditManager{})
			if err != nil {
				require.Equal(t, err, nil)
			}
			res := httptest.NewRecorder()

			body := tt.url
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(body))

			h.PostURLHandler(res, req)

			assert.Equal(t, tt.expectedCode, res.Code, "код ответа не совпадает с ожидаемым")
		})
	}
}

func TestIDGetHandler(t *testing.T) {
	tempFile, err := os.CreateTemp(os.TempDir(), "Iter9")
	if err != nil {
		t.Fatal("failed create temp file", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name()) // удаляем файл после теста
	}()

	path := tempFile.Name()
	url := "http://rambler"
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))

	logger, err := logger.Initialize("Info")
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}

	cfg := &config.Cnfg{ServRunAddr: "", AddrForURL: "", StoragePath: path}
	h, err := NewHandler(cfg, logger, service.URLStorage{}, &audit.AuditManager{})
	if err != nil {
		require.Equal(t, err, nil)
	}

	h.PostURLHandler(res, req)
	assert.Equal(t, http.StatusCreated, res.Code, "код ответа не совпадает с ожидаемым")

	shortURL := res.Body.String()

	tests := []struct {
		name         string // description of this test case
		method       string
		url          string
		expectedCode int
	}{
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

func TestPostShortenHandler(t *testing.T) {
	tempFile, err := os.CreateTemp(os.TempDir(), "Iter9")
	if err != nil {
		t.Fatal("failed create temp file", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name()) // удаляем файл после теста
	}()

	path := tempFile.Name()

	logger, err := logger.Initialize("Info")
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}

	// Создаём тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, err := NewHandler(&config.Cnfg{
			AddrForURL: "http://test-host", StoragePath: path,
		}, logger, service.URLStorage{}, &audit.AuditManager{})
		require.Equal(t, err, nil)
		h.PostURLJSONHandler(w, r)
	}))
	defer ts.Close()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedResult bool
	}{
		{
			name:           "Valid URL",
			requestBody:    `{"url": "https://example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedResult: true,
		},
		{
			name:           "Empty URL",
			requestBody:    `{"url": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectedResult: false,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url": }`,
			expectedStatus: http.StatusBadRequest,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedResult {
				var result map[string]string
				err = json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}
				if result["result"] == "" {
					t.Errorf("expected result %t, got %q", tt.expectedResult, result["result"])
				}
			}
		})
	}
}
