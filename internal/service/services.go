package service

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
)

func InitMap() *model.StringMap {
	return &model.StringMap{
		Data: make(map[string]string),
	}
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
