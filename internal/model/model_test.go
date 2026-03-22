package model

import (
	"fmt"
	"testing"
)

func TestStringMap_GetFullURL(t *testing.T) {
	// Создаём тестовые данные
	testData := TData{
		"key1": DataRow{
			URL:       "https://example.com/1",
			UserID:    "user123",
			IsDeleted: false,
		},
		"key2": DataRow{
			URL:       "https://example.com/2",
			UserID:    "user456",
			IsDeleted: true,
		},
	}

	tests := []struct {
		name    string
		key     string
		want    DataRow
		wantErr bool
		errMsg  string // ожидаемое сообщение об ошибке
	}{
		{
			name:    "успешное получение существующего ключа",
			key:     "key1",
			want:    testData["key1"],
			wantErr: false,
		},
		{
			name:    "получение другого существующего ключа",
			key:     "key2",
			want:    testData["key2"],
			wantErr: false,
		},
		{
			name:    "попытка получить несуществующий ключ",
			key:     "nonexistent",
			want:    DataRow{},
			wantErr: true,
			errMsg:  "id отсутствует",
		},
		{
			name:    "пустой ключ",
			key:     "",
			want:    DataRow{},
			wantErr: true,
			errMsg:  "id пустой",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём и инициализируем StringMap
			m := &StringMap{
				Data: testData,
			}

			got, gotErr := m.GetFullURL(tt.key)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetFullURL() unexpected error: %v", gotErr)
				} else {
					// Проверяем сообщение об ошибке, если оно указано
					if tt.errMsg != "" && gotErr.Error() != tt.errMsg {
						t.Errorf("GetFullURL() error message = %q, want %q", gotErr.Error(), tt.errMsg)
					}
				}
				return
			}

			if tt.wantErr {
				t.Fatal("GetFullURL() succeeded unexpectedly")
			}

			// Сравниваем полученные и ожидаемые значения
			if got != tt.want {
				t.Errorf("GetFullURL() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func BenchmarkStringMap_GetFullURL(b *testing.B) {
	// Подготавливаем тестовые данные заранее, чтобы время инициализации не влияло на замеры
	testData := TData{
		"existing-key": DataRow{
			URL:       "https://example.com/1",
			UserID:    "user123",
			IsDeleted: false,
		},
		"another-key": DataRow{
			URL:       "https://example.com/2",
			UserID:    "user456",
			IsDeleted: true,
		},
		// Добавляем больше ключей для реалистичного теста
		"key-100": DataRow{
			URL:       "https://example.com/100",
			UserID:    "user789",
			IsDeleted: false,
		},
	}

	// Создаём и инициализируем StringMap один раз
	m := &StringMap{
		Data: testData,
	}

	b.ResetTimer() // Сбрасываем таймер, чтобы не учитывать время подготовки данных

	// Тест 1: успешный поиск существующего ключа
	b.Run("existing_key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.GetFullURL("existing-key")
		}
	})

	// Тест 2: поиск другого существующего ключа
	b.Run("another_existing_key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.GetFullURL("another-key")
		}
	})

	// Тест 3: поиск несуществующего ключа (ошибка)
	b.Run("nonexistent_key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.GetFullURL("nonexistent")
		}
	})

	// Тест 4: пустой ключ (ошибка)
	b.Run("empty_key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.GetFullURL("")
		}
	})
}

func TestStringMap_GetAllURLsForUser(t *testing.T) {
	// Подготавливаем тестовые данные
	testData := TData{
		"short1": DataRow{
			URL:       "https://example.com/1",
			UserID:    "user123",
			IsDeleted: false,
		},
		"short2": DataRow{
			URL:       "https://example.com/2",
			UserID:    "user456",
			IsDeleted: true,
		},
		"short3": DataRow{
			URL:       "https://example.com/3",
			UserID:    "user123", // тот же пользователь
			IsDeleted: false,
		},
		"short4": DataRow{
			URL:       "https://example.com/4",
			UserID:    "user789",
			IsDeleted: false,
		},
	}

	tests := []struct {
		name   string
		userID string
		want   map[string]string
	}{
		{
			name:   "пользователь с двумя URL",
			userID: "user123",
			want: map[string]string{
				"short1": "https://example.com/1",
				"short3": "https://example.com/3",
			},
		},
		{
			name:   "пользователь с одним URL",
			userID: "user456",
			want: map[string]string{
				"short2": "https://example.com/2",
			},
		},
		{
			name:   "пользователь без URL",
			userID: "nonexistent-user",
			want:   map[string]string{},
		},
		{
			name:   "пустой userID",
			userID: "",
			want:   map[string]string{},
		},
		{
			name:   "все URL для пользователя с удалёнными записями",
			userID: "user456", // у этого пользователя IsDeleted = true
			want: map[string]string{
				"short2": "https://example.com/2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём и инициализируем StringMap
			m := &StringMap{
				Data: testData,
			}

			got := m.GetAllURLsForUser(tt.userID)

			// Сравниваем длины карт
			if len(got) != len(tt.want) {
				t.Errorf("GetAllURLsForUser() returned %d URLs, want %d", len(got), len(tt.want))
				return
			}

			// Детальное сравнение содержимого карт
			for key, expectedURL := range tt.want {
				actualURL, exists := got[key]
				if !exists {
					t.Errorf("GetAllURLsForUser() missing key %q", key)
					continue
				}
				if actualURL != expectedURL {
					t.Errorf("GetAllURLsForUser()[%q] = %q, want %q", key, actualURL, expectedURL)
				}
			}

			// Проверяем, что в результате нет лишних ключей
			for key := range got {
				if _, exists := tt.want[key]; !exists {
					t.Errorf("GetAllURLsForUser() contains unexpected key %q", key)
				}
			}
		})
	}
}

func BenchmarkStringMap_GetAllURLsForUser(b *testing.B) {
	// Подготавливаем тестовые данные разного размера
	smallData := generateTestData(10)    // 10 записей
	mediumData := generateTestData(1000) // 1 000 записей
	largeData := generateTestData(10000) // 10 000 записей

	tests := []struct {
		name   string
		m      *StringMap
		userID string
	}{
		{
			name:   "малый набор данных, пользователь с 2 URL",
			m:      &StringMap{Data: smallData},
			userID: "user-0",
		},
		{
			name:   "средний набор данных, пользователь с 50 URL",
			m:      &StringMap{Data: mediumData},
			userID: "user-5",
		},
		{
			name:   "большой набор данных, пользователь с 100 URL",
			m:      &StringMap{Data: largeData},
			userID: "user-10",
		},
		{
			name:   "большой набор данных, пользователь без URL",
			m:      &StringMap{Data: largeData},
			userID: "nonexistent-user",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer() // Сбрасываем таймер, чтобы не учитывать подготовку данных

			for i := 0; i < b.N; i++ {
				tt.m.GetAllURLsForUser(tt.userID)
			}
		})
	}
}

// Вспомогательная функция для генерации тестовых данных
func generateTestData(size int) TData {
	data := make(TData, size)
	users := []string{"user-0", "user-1", "user-2", "user-3", "user-4", "user-5"}

	for i := 0; i < size; i++ {
		userIndex := i % len(users)
		data[fmt.Sprintf("short-%d", i)] = DataRow{
			URL:       fmt.Sprintf("https://example.com/%d", i),
			UserID:    users[userIndex],
			IsDeleted: i%10 == 0, // каждая 10‑я запись — удалённая
		}
	}

	return data
}
