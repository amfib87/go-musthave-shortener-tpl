package service

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/repository"
)

const maxRetries = 5
const CookieName = "user_auth"

func InitMap(st URLStorage) (*model.StringMap, error) {
	stringMap := &model.StringMap{
		Data: make(model.TData),
	}

	if st.File != nil {

		reader := bufio.NewReader(st.File)
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed read data from file: %w", err)
		}

		if len(data) != 0 {
			if err := json.Unmarshal(data, &stringMap.Data); err != nil {
				return nil, fmt.Errorf("failed json Unmarshal: %v", err)
			}
		}

	} else if st.DB != nil {
		dataDB, err := model.ReadDB(st.DB)
		if err != nil {
			return nil, fmt.Errorf("failed read DB: %v", err)
		}
		stringMap.Data = dataDB
	}

	return stringMap, nil
}

func GetShortURL(ctx context.Context, data model.DataRow, m *model.StringMap, st URLStorage, lg *logger.TLog) (string, error) {
	const maxRetries = 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortURL := generateShortID()

		shortURLExist, err := m.InsertShortURL(ctx, data, shortURL, st.File, st.DB)
		if err != nil {
			lg.Lg.Sugar().Infoln("error with InsertShortURL:", err.Error())
		}

		if errors.Is(err, model.ErrOriginalURLExist) {
			return shortURLExist, err
		}
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

func GetShortURLMass(ctx context.Context, values []model.DataRequestMass, m *model.StringMap, st URLStorage, userID string) ([]model.DataAnswerMass, error) {
	export := []model.DataAnswerMass{}
	shortKeys := make(model.TData)

	for _, lineData := range values {
		for attempt := 0; attempt < maxRetries; attempt++ {
			shortURL := generateShortID()

			// проверяем на наличие сгенерированного shorturl
			count, err := model.CheckExistShortURL(ctx, shortURL, st.DB)
			if err != nil {
				return nil, fmt.Errorf("CheckExistShortURL, err: %w", err)
			}
			if count != 0 {
				continue
			}

			shortKeys[shortURL] = model.DataRow{
				URL:    lineData.OriginalURL,
				UserID: userID}
			export = append(export, model.DataAnswerMass{CorrelationID: lineData.CorrelationID, ShortURL: shortURL})
			break
		}
	}

	if len(export) == 0 {
		return nil, fmt.Errorf("failed to compose unique short URL after %d attempts", maxRetries)
	}

	if err := m.InsertShortURLMass(ctx, shortKeys, st.DB, st.File); err != nil {
		return nil, fmt.Errorf("failed insertShortURLMass, err: %w", err)
	}

	return export, nil
}

type URLStorage struct {
	DB   *sql.DB
	File *os.File
}

func InitURLStorage(cfg *config.Cnfg, log *logger.TLog) (URLStorage, error) {
	URLstorage := URLStorage{}
	var err error

	if cfg.DataBaseDsn != "" {
		URLstorage.DB, err = repository.InitDB(cfg.DataBaseDsn)
		if err != nil {
			log.Lg.Sugar().Fatalf("failed InitDB: %v", err)
			return URLstorage, err
		}

	} else if cfg.StoragePath != "" {
		URLstorage.File, err = InitFile(cfg.StoragePath)
		if err != nil {
			log.Lg.Sugar().Fatalf("failed to init file: %v", err)
			return URLstorage, err
		}
	}

	return URLstorage, nil
}

func (st URLStorage) Close(log *logger.TLog) {
	if st.DB != nil {
		defer st.DB.Close()
	}
	if st.File != nil {
		defer FileClose(st.File, log)
	}

}

func DelShortURLs(shortURLs []model.ShortURL, userID string, st URLStorage, data *model.StringMap) error {
	const batchSize = 10
	ch := make(chan []string, batchSize)

	go func() {
		defer close(ch)
		var ids []string
		for _, shortURL := range shortURLs {
			ids = append(ids, string(shortURL)) // предполагается, что поле называется ShortID
		}

		for i := 0; i < len(ids); i += batchSize {
			end := i + batchSize
			if end > len(ids) {
				end = len(ids)
			}
			ch <- ids[i:end]
		}
	}()

	var wg sync.WaitGroup
	for batch := range ch {
		wg.Add(1)
		go func(b []string) {
			defer wg.Done()
			if err := model.MarkAsDeleted(b, userID, st.DB, data); err != nil {
				log.Printf("Failed to mark batch as deleted: %v", err)
			}
		}(batch)
	}
	wg.Wait()
	return nil
}
