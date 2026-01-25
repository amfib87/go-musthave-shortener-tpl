package model

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type TData map[string]string

var ErrKeyExists = errors.New("key already exists")
var ErrOriginalURLExist = errors.New("original url exists")

type StringMap struct {
	mu   sync.Mutex
	Data TData `json:"data"`
}

// Структура входного элемента
type DataRequestMass struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// Структура выходного элемента
type DataAnswerMass struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (m *StringMap) InsertShortURL(ctx context.Context, key, shortURL string, file *os.File, db *sql.DB) (shortURLExist string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := (m.Data)[shortURL]; exists {
		return "", ErrKeyExists
	}

	if db != nil {
		err := saveToDB(ctx, shortURL, key, db)
		if errors.Is(err, ErrOriginalURLExist) {

			shortURLExist, err = getExistShortURL(ctx, key, db)
			if err != nil {
				return "", fmt.Errorf("failed getExistShortURL: %w", err)
			}
			return shortURLExist, ErrOriginalURLExist

		} else if err != nil {
			return "", err
		}

	} else if file != nil {

		if err := saveFile(m.Data, file); err != nil {
			return "", err
		}
	}

	(m.Data)[shortURL] = key

	return "", nil
}

func (m *StringMap) InsertShortURLMass(ctx context.Context, values map[string]string, db *sql.DB, file *os.File) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for short, full := range values {
		if _, exists := (m.Data)[short]; exists {
			return ErrKeyExists
		}
		(m.Data)[short] = full
	}

	if db != nil {
		for short, full := range values {
			if err := saveToDB(ctx, short, full, db); err != nil {
				return err
			}
		}
	} else if file != nil {

		if err := saveFile(m.Data, file); err != nil {
			return err
		}
	}

	return nil
}

func (m *StringMap) GetFullURL(key string) (val string, err error) {
	value, ok := (m.Data)[key]
	if !ok {
		return "", fmt.Errorf("id отсутствует")
	}
	return value, nil
}

func saveFile(data TData, f *os.File) error {
	// сериализуем структуру
	dataJSON, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)

	// сохраняем данные в файл
	if _, err := writer.Write(dataJSON); err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func saveToDB(ctx context.Context, shortURL, key string, db *sql.DB) error {
	query := `INSERT INTO tdata (shorturl, originalurl) VALUES ($1, $2) ON CONFLICT (originalurl) DO NOTHING RETURNING id`
	var newID int
	err := db.QueryRowContext(ctx, query, shortURL, key).Scan(&newID)

	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("url already exists: %s: %w", shortURL, ErrOriginalURLExist)

	case err != nil:
		return fmt.Errorf("failed to insert URL: %w", err)

	case newID == 0:
		return fmt.Errorf("line id not recieved during insert")

	default:
		return nil
	}
}

func ReadDB(db *sql.DB) (data TData, err error) {
	query := `SELECT shorturl, originalurl FROM tdata`
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data = make(TData)

	// пробегаем по всем записям
	for rows.Next() {
		var shortURL string
		var fullURL string
		err = rows.Scan(&shortURL, &fullURL)
		if err != nil {
			return nil, err
		}
		data[shortURL] = fullURL
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getExistShortURL(ctx context.Context, originalURL string, db *sql.DB) (shortURL string, err error) {
	querySel := `SELECT shorturl FROM tdata WHERE originalurl = $1`
	row := db.QueryRowContext(ctx, querySel, originalURL)

	var shortURLExist string
	err = row.Scan(&shortURLExist)
	if err != nil {
		return "", fmt.Errorf("failed scan short url %w", err)
	}

	return shortURLExist, nil
}

func CheckExistShortURL(ctx context.Context, shortURL string, db *sql.DB) (int, error) {
	var count int
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) as count FROM tdata WHERE shorturl = $1", shortURL)
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed queryrow, err: %w", err)
	}
	return count, nil
}
