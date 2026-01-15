package model

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
)

type TData map[string]string

var ErrKeyExists = errors.New("key already exists")

type StringMap struct {
	mu   sync.Mutex
	Data TData `json:"data"`
}

func (m *StringMap) InsertShortURL(key, shortURL string, f *os.File, bd *sql.DB, ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := (m.Data)[shortURL]; exists {
		return ErrKeyExists
	}

	(m.Data)[shortURL] = key
	if bd != nil {
		if err := saveToBD(shortURL, key, bd, ctx); err != nil {
			return err
		}
	} else if f != nil {

		if err := saveFile(m.Data, f); err != nil {
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

func saveToBD(shortURL, key string, bd *sql.DB, ctx context.Context) error {
	query := `INSERT INTO tdata (shorturl, originalurl) VALUES ($1, $2)`
	_, err := bd.ExecContext(ctx, query, shortURL, key)
	if err != nil {
		// Выводим детальную информацию об ошибке
		log.Printf("SQL Error: %v", err)
		log.Printf("shortURL: %q, key: %q", shortURL, key)
	}
	return err
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
