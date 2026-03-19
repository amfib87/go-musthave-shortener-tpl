package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewAuditEvent(t *testing.T) {
	tests := []struct {
		name   string
		action string
		userID string
		url    string
		want   func() *AuditEvent // используем функцию, чтобы корректно сравнивать время
	}{
		{
			name:   "basic event creation",
			action: "shorten",
			userID: "12315134",
			url:    "https://mylongdomain.com/my/long/path/to/shorten/",
			want: func() *AuditEvent {
				return &AuditEvent{
					Ts:     time.Now().Unix(),
					Action: "shorten",
					UserID: "12315134",
					URL:    "https://mylongdomain.com/my/long/path/to/shorten/",
				}
			},
		},
		{
			name:   "follow action",
			action: "follow",
			userID: "",
			url:    "https://example.com/abc123",
			want: func() *AuditEvent {
				return &AuditEvent{
					Ts:     time.Now().Unix(),
					Action: "follow",
					UserID: "",
					URL:    "https://example.com/abc123",
				}
			},
		},
		{
			name:   "empty values",
			action: "",
			userID: "",
			url:    "",
			want: func() *AuditEvent {
				return &AuditEvent{
					Ts:     time.Now().Unix(),
					Action: "",
					UserID: "",
					URL:    "",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuditEvent(tt.action, tt.userID, tt.url)
			want := tt.want()

			// Сравниваем все поля, кроме Ts (времени)
			if got.Action != want.Action {
				t.Errorf("Action = %v, want %v", got.Action, want.Action)
			}
			if got.UserID != want.UserID {
				t.Errorf("UserID = %v, want %v", got.UserID, want.UserID)
			}
			if got.URL != want.URL {
				t.Errorf("URL = %v, want %v", got.URL, want.URL)
			}

			// Проверяем, что Ts — это текущее время (в пределах ±1 секунды)
			now := time.Now().Unix()
			if got.Ts < now-1 || got.Ts > now+1 {
				t.Errorf("Ts = %v, expected to be close to current time %v", got.Ts, now)
			}
		})
	}
}

func TestNewFileAuditSubscriber(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		cleanup  func() // функция для очистки после теста
	}{
		{
			name:     "empty file path (audit disabled)",
			filePath: "",
			wantErr:  false,
		},
		{
			name:     "valid file path - new file",
			filePath: "test_audit_1.log",
			wantErr:  false,
			cleanup: func() {
				os.Remove("test_audit_1.log")
			},
		},
		{
			name:     "valid file path - existing file",
			filePath: "test_audit_2.log",
			wantErr:  false,
			cleanup: func() {
				os.Remove("test_audit_2.log")
			},
		},
		{
			name:     "invalid path - directory doesn't exist",
			filePath: "/nonexistent/path/audit.log",
			wantErr:  true,
		},
		{
			name:     "invalid path - permission denied",
			filePath: "/root/restricted.log", // на большинстве систем нет доступа
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewFileAuditSubscriber(tt.filePath)

			// Проверка наличия ошибки
			if (gotErr != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("NewFileAuditSubscriber() expected error, but got nil")
				} else {
					t.Errorf("NewFileAuditSubscriber() unexpected error: %v", gotErr)
				}
				return
			}

			// Если ожидается ошибка, проверяем только её наличие
			if tt.wantErr {
				if got != nil {
					t.Error("NewFileAuditSubscriber() should return nil when error occurs")
				}
				return
			}

			// Если ошибки нет, проверяем корректность результата
			if got == nil && tt.wantErr {
				t.Fatal("NewFileAuditSubscriber() returned nil, but no error occurred")
			}

			// Проверяем, что файл действительно открыт
			if tt.filePath != "" && got.file == nil {
				t.Error("FileAuditSubscriber.file is nil, but file should be opened")
			}

			if tt.filePath != "" {
				// Дополнительная проверка: пытаемся записать тестовое сообщение
				testMsg := []byte("test audit message\n")
				_, err := got.file.Write(testMsg)
				if err != nil {
					t.Errorf("Cannot write to opened file: %v", err)
				}

				// Закрываем файл после теста, чтобы избежать утечек
				if err := got.file.Close(); err != nil && !tt.wantErr {
					t.Errorf("Error closing file after test: %v", err)
				}

				// Выполняем очистку, если она определена
				if tt.cleanup != nil {
					tt.cleanup()
				}
			}
		})
	}
}

func TestFileAuditSubscriber_Notify(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		event    *AuditEvent
		wantErr  bool
		cleanup  func() // функция для очистки после теста
	}{
		{
			name:     "successful write to file",
			filePath: "test_audit_notify_1.log",
			event: &AuditEvent{
				Ts:     12345678,
				Action: "shorten",
				UserID: "12315134",
				URL:    "https://mylongdomain.com/my/long/path/to/shorten/",
			},
			wantErr: false,
			cleanup: func() {
				os.Remove("test_audit_notify_1.log")
			},
		},
		{
			name:     "nil subscriber",
			filePath: "", // вернёт nil subscriber
			event: &AuditEvent{
				Ts:     12345678,
				Action: "follow",
				UserID: "",
				URL:    "https://example.com/abc123",
			},
			wantErr: true,
		},
		{
			name:     "empty event fields",
			filePath: "test_audit_notify_2.log",
			event: &AuditEvent{
				Ts:     99999999,
				Action: "",
				UserID: "",
				URL:    "",
			},
			wantErr: false,
			cleanup: func() {
				os.Remove("test_audit_notify_2.log")
			},
		},
		{
			name:     "large event data",
			filePath: "test_audit_notify_3.log",
			event: &AuditEvent{
				Ts:     1234567890,
				Action: "shorten",
				UserID: "987654321",
				URL:    "https://very.long.domain.name/with/a/very/long/path/that/exceeds/typical/length/limits",
			},
			wantErr: false,
			cleanup: func() {
				os.Remove("test_audit_notify_3.log")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очистка файла перед тестом (гарантируем пустой файл)
			if tt.filePath != "" {
				_ = os.Remove(tt.filePath) // игнорируем ошибку, если файла не было
			}

			f, err := NewFileAuditSubscriber(tt.filePath)

			if tt.filePath == "" {
				if err != nil {
					t.Fatalf("NewFileAuditSubscriber() unexpected error for empty path: %v", err)
				}
				if f != nil {
					t.Fatal("NewFileAuditSubscriber() should return nil for empty filePath")
				}

				var nilSubscriber *FileAuditSubscriber
				gotErr := nilSubscriber.Notify(tt.event)
				if !tt.wantErr {
					t.Errorf("Notify() on nil subscriber expected no error, but got: %v", gotErr)
				}
				return
			} else {
				if err != nil {
					t.Fatalf("could not construct receiver type: %v", err)
				}
				defer f.file.Close()
			}

			gotErr := f.Notify(tt.event)

			// Проверка наличия ошибки
			if (gotErr != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("Notify() expected error, but got nil")
				} else {
					t.Errorf("Notify() unexpected error: %v", gotErr)
				}
				return
			}

			// Если ошибки нет и файл указан, проверяем содержимое
			if !tt.wantErr && tt.filePath != "" {
				data, readErr := os.ReadFile(tt.filePath)
				if readErr != nil {
					t.Errorf("cannot read audit file after write: %v", readErr)
					return
				}

				expectedJSON, marshalErr := json.Marshal(tt.event)
				if marshalErr != nil {
					t.Fatalf("cannot marshal expected event: %v", marshalErr)
				}
				expectedLine := string(expectedJSON) + "\n"

				if string(data) != expectedLine {
					t.Errorf(
						"file content mismatch\ngot:  %q\nwant: %q",
						string(data),
						expectedLine,
					)
				}
			}

			// Выполняем очистку, если она определена
			if tt.cleanup != nil {
				tt.cleanup()
			}
		})
	}
}

func TestNewRemoteAuditSubscriber(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL provided",
			url:     "http://localhost:8080/audit",
			wantErr: false,
		},
		{
			name:    "empty URL (subscriber disabled)",
			url:     "",
			wantErr: false,
		},
		{
			name:    "URL with path",
			url:     "https://api.example.com/v1/audit-events",
			wantErr: false,
		},
		{
			name:    "IP address URL",
			url:     "http://192.168.1.100:3000/log",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewRemoteAuditSubscriber(tt.url)

			// Проверка наличия ошибки
			if (gotErr != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("NewRemoteAuditSubscriber() expected error, but got nil")
				} else {
					t.Errorf("NewRemoteAuditSubscriber() unexpected error: %v", gotErr)
				}
				return
			}

			// Если ожидается ошибка, проверяем только её наличие
			if tt.wantErr {
				if got != nil {
					t.Error("NewRemoteAuditSubscriber() should return nil when error occurs")
				}
				return
			}

			// Если ошибки нет, проверяем корректность результата
			if got == nil && tt.url != "" {
				t.Fatal("NewRemoteAuditSubscriber() returned nil, but no error occurred")
			}

			if tt.url != "" {
				// Проверяем, что URL установлен корректно
				if got.url != tt.url {
					t.Errorf("RemoteAuditSubscriber.url = %q, want %q", got.url, tt.url)
				}

				// Проверяем, что клиент создан и имеет правильный таймаут
				if got.client == nil {
					t.Error("RemoteAuditSubscriber.client is nil, but should be initialized")
				} else {
					if got.client.Timeout != 5*time.Second {
						t.Errorf("http.Client.Timeout = %v, want %v", got.client.Timeout, 5*time.Second)
					}
				}
			}
		})
	}
}

func TestRemoteAuditSubscriber_Notify(t *testing.T) {
	// Запускаем тестовый HTTP‑сервер для имитации удалённого аудита
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос имеет правильный Content‑Type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		// Читаем тело запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusInternalServerError)
			return
		}

		// Проверяем валидность JSON
		var event AuditEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Возвращаем успешный статус
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		event   *AuditEvent
		wantErr bool
	}{
		{
			name: "successful notification - valid event",
			url:  server.URL,
			event: &AuditEvent{
				Ts:     12345678,
				Action: "shorten",
				UserID: "12315134",
				URL:    "https://mylongdomain.com/my/long/path/to/shorten/",
			},
			wantErr: false,
		},
		{
			name: "nil subscriber",
			url:  "", // вернёт nil subscriber
			event: &AuditEvent{
				Ts:     12345678,
				Action: "follow",
				UserID: "",
				URL:    "https://example.com/abc123",
			},
			wantErr: true,
		},
		{
			name: "server returns error status",
			url: func() string {
				// Сервер, который всегда возвращает ошибку
				errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				t.Cleanup(errorServer.Close)
				return errorServer.URL
			}(),
			event: &AuditEvent{
				Ts:     99999999,
				Action: "test",
				UserID: "testuser",
				URL:    "https://test.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRemoteAuditSubscriber(tt.url)

			// Для случая с пустым URL ожидается nil subscriber и отсутствие ошибки
			if tt.url == "" {
				if err != nil {
					t.Fatalf("NewRemoteAuditSubscriber() unexpected error for empty URL: %v", err)
				}
				if r != nil {
					t.Fatal("NewRemoteAuditSubscriber() should return nil for empty URL")
				}

				// Тестируем поведение при nil subscriber
				var nilSubscriber *RemoteAuditSubscriber
				gotErr := nilSubscriber.Notify(tt.event)
				if !tt.wantErr {
					t.Errorf("Notify() on nil subscriber expected no error, but got: %v", gotErr)
				}
				return
			} else {
				if err != nil {
					t.Fatalf("could not construct receiver type: %v", err)
				}
			}

			gotErr := r.Notify(tt.event)

			// Проверка наличия ошибки
			if (gotErr != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("Notify() expected error, but got nil")
				} else {
					t.Errorf("Notify() unexpected error: %v", gotErr)
				}
				return
			}
		})
	}
}

func TestNewAuditManager(t *testing.T) {
	tests := []struct {
		name                string
		filePath            string
		URL                 string
		wantErr             bool
		expectedSubscribers int // ожидаемое количество подписчиков
	}{
		{
			name:                "both subscribers enabled",
			filePath:            "test_file.log",
			URL:                 "http://localhost:8080/audit",
			wantErr:             false,
			expectedSubscribers: 2,
		},
		{
			name:                "only file subscriber enabled",
			filePath:            "test_file.log",
			URL:                 "",
			wantErr:             false,
			expectedSubscribers: 1,
		},
		{
			name:                "only remote subscriber enabled",
			filePath:            "",
			URL:                 "http://localhost:8080/audit",
			wantErr:             false,
			expectedSubscribers: 1,
		},
		{
			name:                "no subscribers enabled",
			filePath:            "",
			URL:                 "",
			wantErr:             false,
			expectedSubscribers: 0,
		},
		{
			name:     "invalid file path",
			filePath: "/invalid/path/test.log", // может вызвать ошибку ОС
			URL:      "http://localhost:8080/audit",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewAuditManager(tt.filePath, tt.URL)

			// Проверка наличия ошибки
			if (gotErr != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("NewAuditManager() expected error, but got nil")
				} else {
					t.Errorf("NewAuditManager() unexpected error: %v", gotErr)
				}
				return
			}

			// Если ожидается ошибка, проверяем только её наличие
			if tt.wantErr {
				if got != nil {
					t.Error("NewAuditManager() should return nil when error occurs")
				}
				return
			}

			// Если ошибки нет, проверяем корректность результата
			if got == nil {
				t.Fatal("NewAuditManager() returned nil, but no error occurred")
			}

			// Проверяем количество подписчиков
			if len(got.subscribers) != tt.expectedSubscribers {
				t.Errorf(
					"AuditManager has %d subscribers, want %d",
					len(got.subscribers),
					tt.expectedSubscribers,
				)
			}

			// Дополнительная проверка типов подписчиков (опционально)
			if !tt.wantErr && len(got.subscribers) > 0 {
				for i, sub := range got.subscribers {
					switch sub.(type) {
					case *FileAuditSubscriber:
						if tt.filePath == "" {
							t.Errorf("subscriber %d is FileAuditSubscriber, but filePath was empty", i)
						}
					case *RemoteAuditSubscriber:
						if tt.URL == "" {
							t.Errorf("subscriber %d is RemoteAuditSubscriber, but URL was empty", i)
						}
					default:
						t.Errorf("unknown subscriber type at index %d", i)
					}
				}
			}
		})
	}
}
