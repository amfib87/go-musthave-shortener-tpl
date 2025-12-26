package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
)

func InitMap(name string) (*model.StringMap, error) {
	if name == "" {
		return nil, os.ErrInvalid
	}

	var data []byte

	_, err := os.Stat(name)
	if err == nil {
		data, err = os.ReadFile(name)
		if err != nil {
			return nil, err
		}
	}

	stringMap := &model.StringMap{
		Data: make(model.TData),
		Name: name,
	}

	if len(data) != 0 {
		if err := json.Unmarshal(data, &stringMap.Data); err != nil {
			return nil, err
		}
	}

	return stringMap, nil
}

func GetShortURL(key string, m *model.StringMap) (string, error) {
	const maxRetries = 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortURL := generateShortID()

		err := m.InsertShortURL(key, shortURL)
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
