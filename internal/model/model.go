// Package model предназначен для реализации логики работы хранения и обработки данных
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

	"github.com/lib/pq"
)

type TData map[string]DataRow

var ErrKeyExists = errors.New("key already exists")
var ErrOriginalURLExist = errors.New("original url exists")

type StringMap struct {
	mu   sync.Mutex
	Data TData `json:"data"`
}

// DataRequestMass - Структура входного элемента
type DataRequestMass struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// DataAnswerMass - Структура выходного элемента
type DataAnswerMass struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type DataRow struct {
	URL       string `json:"url"`
	UserID    string `json:"userid"`
	IsDeleted bool   `json:"-"`
}

type AllURLAnswer struct {
	ShortURL string `json:"short_url"`
	OrigURL  string `json:"original_url"`
}

type ContextKey string

type ShortURL string

const SecretKey = "secret_key"

func (m *StringMap) InsertShortURL(ctx context.Context, data DataRow, shortURL string, file *os.File, db *sql.DB) (shortURLExist string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := (m.Data)[shortURL]; exists {
		return "", ErrKeyExists
	}

	if db != nil {
		err := saveToDB(ctx, shortURL, data, db)
		if errors.Is(err, ErrOriginalURLExist) {

			shortURLExist, err = getExistShortURL(ctx, data.URL, db)
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

	(m.Data)[shortURL] = data

	return "", nil
}

func (m *StringMap) InsertShortURLMass(ctx context.Context, values TData, db *sql.DB, file *os.File) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for short, data := range values {
		if _, exists := (m.Data)[short]; exists {
			return ErrKeyExists
		}
		(m.Data)[short] = data
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

func (m *StringMap) GetFullURL(key string) (DataRow, error) {
	if key == "" {
		return DataRow{}, fmt.Errorf("id пустой")
	}

	value, ok := (m.Data)[key]
	if !ok {
		return DataRow{}, fmt.Errorf("id отсутствует")
	}
	return value, nil
}

func (m *StringMap) GetAllURLsForUser(userID string) map[string]string {
	allURLs := make(map[string]string)

	for short, data := range m.Data {
		if data.UserID == userID {
			allURLs[short] = data.URL
		}
	}

	return allURLs
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
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}

func saveToDB(ctx context.Context, shortURL string, data DataRow, db *sql.DB) error {
	query := `INSERT INTO tdata (shorturl, originalurl, userID) VALUES ($1, $2, $3) ON CONFLICT (originalurl) DO NOTHING RETURNING id`
	var newID int
	err := db.QueryRowContext(ctx, query, shortURL, data.URL, data.UserID).Scan(&newID)

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
	query := `SELECT shorturl, originalurl, userID, is_deleted FROM tdata`
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	data = make(TData)

	// пробегаем по всем записям
	for rows.Next() {
		var (
			shortURL  string
			fullURL   string
			userID    string
			isDeleted bool
		)

		err = rows.Scan(&shortURL, &fullURL, &userID, &isDeleted)
		if err != nil {
			return nil, err
		}

		data[shortURL] = DataRow{
			URL:       fullURL,
			UserID:    userID,
			IsDeleted: isDeleted}
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
	if db != nil {
		row := db.QueryRowContext(ctx, "SELECT COUNT(*) as count FROM tdata WHERE shorturl = $1", shortURL)
		err := row.Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed queryrow, err: %w", err)
		}
	}
	return count, nil
}

func MarkAsDeleted(shortURLs []string, userID string, db *sql.DB, data *StringMap) error {
	query := `UPDATE tdata SET is_deleted = TRUE WHERE shorturl = ANY($1) AND userid = $2`
	_, err := db.ExecContext(context.Background(), query, pq.Array(shortURLs), userID)
	if err != nil {
		return err
	}

	for _, shortURL := range shortURLs {
		dataRow, ok := data.Data[shortURL]
		if !ok {
			continue
		}

		if dataRow.UserID == userID {
			dataRow.IsDeleted = true
			data.Data[shortURL] = dataRow
		}
	}

	return nil
}
