package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
)

func InitMap(file *os.File) (*model.StringMap, error) {
	var data []byte

	reader := bufio.NewReader(file)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed read data from file: %w", err)
	}

	stringMap := &model.StringMap{
		Data: make(model.TData),
	}

	if len(data) != 0 {
		if err := json.Unmarshal(data, &stringMap.Data); err != nil {
			return nil, fmt.Errorf("failed Unmarshal: %v", err)
		}
	}

	return stringMap, nil
}

func GetShortURL(key string, m *model.StringMap, f *os.File) (string, error) {
	const maxRetries = 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortURL := generateShortID()

		err := m.InsertShortURL(key, shortURL, f)
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
