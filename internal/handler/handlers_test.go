package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/audit"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testUser string = "test-user"
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

// ExampleHandler_PostURLHandler демонстрирует использование PostURLHandler для сокращения URL.
//
// Пример показывает:
//   - успешный случай создания сокращённого URL;
func ExampleHandler_PostURLHandler() {
	// Создаём тестовый сервер и мок‑зависимости
	handler := &Handler{
		Logger: &logger.TLog{Lg: zap.NewExample()},
		cfg:    &config.Cnfg{AddrForURL: "https://short.example.com"},
		mapURL: &model.StringMap{Data: make(model.TData)},
		urlSt:  service.URLStorage{},
		audit:  nil,
	}

	// Мок‑реализация сервиса для тестирования
	originalURL := "https://example.com/very/long/url"

	// Тест 1: Успешное создание сокращённого URL
	{
		req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(originalURL))
		// Добавляем UserID в контекст
		ctx := context.WithValue(req.Context(), userIDKey, testUser)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.PostURLHandler(w, req)

		resp := w.Result()

		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Headers: %v\n", resp.Header)
		// Output:
		// Status: 201
		// Headers: map[Content-Type:[text/plain]]
	}
}

// ExampleHandler_IDGetHandler демонстрирует использование IDGetHandler для перенаправления по сокращённому ID.
//
// Пример показывает:
//   - успешное перенаправление по валидному ID;
func ExampleHandler_IDGetHandler() {
	// Создаём тестовый сервер и мок‑зависимости
	handler := &Handler{
		Logger: &logger.TLog{Lg: zap.NewExample()},
		mapURL: &model.StringMap{Data: make(model.TData)},
		audit:  nil,
	}
	handler.mapURL.Data["abc123"] = model.DataRow{
		URL:    "https://example.com/very/long/url",
		UserID: testUser,
	}

	// Тест 1: Успешное выполнение
	{
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		// Добавляем UserID в контекст
		ctx := context.WithValue(req.Context(), userIDKey, testUser)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.IDGetHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Location: %s\n", resp.Header.Get("Location"))
		fmt.Printf("Headers: %v\n", resp.Header)
		// Output:
		// Status: 307
		// Location: https://example.com/very/long/url
		// Headers: map[Location:[https://example.com/very/long/url]]
	}
}

// ExampleHandler_PostURLJSONHandler демонстрирует использование PostURLJSONHandler для сокращения URL через JSON API.
//
// Пример показывает:
//   - успешный случай создания сокращённого URL;
func ExampleHandler_PostURLJSONHandler() {
	// Создаём тестовый сервер и мок‑зависимости
	handler := &Handler{
		Logger: &logger.TLog{Lg: zap.NewExample()}, // используем простой логгер для примера
		cfg:    &config.Cnfg{AddrForURL: "https://short.example.com"},
		mapURL: &model.StringMap{Data: make(model.TData)},
		audit:  nil,
	}

	// Тест 1: Успешное создание сокращённого URL
	{
		requestBody := `{"url": "https://example.com/very/long/url"}`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		// Добавляем UserID в контекст
		ctx := context.WithValue(req.Context(), userIDKey, testUser)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.PostURLJSONHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Headers: %v\n", resp.Header)
		// Output:
		// Status: 201
		// Headers: map[Content-Type:[application/json]]
	}

}

// ExampleHandler_PostMassURLHandler демонстрирует использование PostMassURLHandler для массового сокращения URL.
//
// Пример показывает:
//   - успешный случай массового сокращения нескольких URL;
func ExampleHandler_PostMassURLHandler() {
	// Создаём тестовый сервер и мок‑зависимости
	handler := &Handler{
		Logger: &logger.TLog{Lg: zap.NewExample()}, // используем простой логгер для примера
		cfg:    &config.Cnfg{AddrForURL: "https://short.example.com"},
		mapURL: &model.StringMap{Data: make(model.TData)},
	}

	// Тест 1: Успешное массовое сокращение URL
	{
		requestBody := `[
		{"url": "https://example.com/page1", "correlation_id": "req-1"},
		{"url": "https://example.com/page2", "correlation_id": "req-2"}
	]`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/mass", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		// Добавляем UserID в контекст
		ctx := context.WithValue(req.Context(), userIDKey, "test-user-123")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.PostMassURLHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Headers: %v\n", resp.Header)
		// Output:
		// Status: 201
		// Headers: map[Content-Type:[application/json]]
	}

}
