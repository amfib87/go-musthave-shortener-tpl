package service

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
)

const maxRetries = 5

func InitMap(file *os.File, db *sql.DB) (*model.StringMap, error) {
	stringMap := &model.StringMap{
		Data: make(model.TData),
	}

	if file != nil {

		reader := bufio.NewReader(file)
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed read data from file: %w", err)
		}

		if len(data) != 0 {
			if err := json.Unmarshal(data, &stringMap.Data); err != nil {
				return nil, fmt.Errorf("failed json Unmarshal: %v", err)
			}
		}

	} else if db != nil {
		dataDB, err := model.ReadDB(db)
		if err != nil {
			return nil, fmt.Errorf("failed read DB: %v", err)
		}
		stringMap.Data = dataDB
	}

	return stringMap, nil
}

func GetShortURL(key string, m *model.StringMap, f *os.File, db *sql.DB, ctx context.Context) (string, error) {
	const maxRetries = 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortURL := generateShortID()

		err := m.InsertShortURL(key, shortURL, f, db, ctx)
		if err == nil {
			return shortURL, nil
		}

		if errors.Is(err, model.ErrKeyExists) {
			continue
		}

		return "", fmt.Errorf("unexpected error on insert")
	}

	return "", fmt.Errorf("failed to compose unique short URL after %d attempts", maxRetries)
}

func generateShortID() string {
	const Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 8)
	for i := range b {
		b[i] = Letters[rand.Intn(len(Letters))]
	}
	return string(b)
}

func InitFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func FileClose(file *os.File, lg *logger.TLog) {
	if err := file.Close(); err != nil {
		lg.Lg.Sugar().Infoln("failed close file: %v", err)
	}
}

func GetShortURLMass(values []model.DataRequestMass, m *model.StringMap, f *os.File, db *sql.DB, ctx context.Context) ([]model.DataAnswerMass, error) {
	export := []model.DataAnswerMass{}
	shortKeys := make(map[string]string)

	for _, lineData := range values {
		for attempt := 0; attempt < maxRetries; attempt++ {
			shortURL := generateShortID()

			// проверяем на наличие сгенерированного shorturl
			var count int
			row := db.QueryRowContext(ctx, "SELECT COUNT(*) as count FROM tdata WHERE shorturl = $1", shortURL)
			err := row.Scan(&count)
			if err != nil {
				return nil, fmt.Errorf("failed queryrow, err: %w", err)
			}
			if count != 0 {
				continue
			}

			shortKeys[shortURL] = lineData.OriginalURL
			export = append(export, model.DataAnswerMass{CorrelationID: lineData.CorrelationID, ShortURL: shortURL})
			break
		}
	}

	if len(export) == 0 {
		return nil, fmt.Errorf("failed to compose unique short URL after %d attempts", maxRetries)
	}

	if err := m.InsertShortURLMass(shortKeys, f, db, ctx); err != nil {
		return nil, fmt.Errorf("failed insertShortURLMass, err: %w", err)
	}

	return export, nil
}
