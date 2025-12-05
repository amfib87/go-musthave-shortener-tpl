package service

import (
	"fmt"
	"math/rand"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
)

func InitMap() *model.StringMap {
	m := make(model.StringMap)
	return &m
}

func GetShortURL(key string, m *model.StringMap) string {
	shortURL := generateShortID()
	_, exists := (*m)[shortURL]
	if !exists {
		(*m)[shortURL] = key
	}

	return shortURL
}

func generateShortID() string {
	const Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 8)
	for i := range b {
		b[i] = Letters[rand.Intn(len(Letters))]
	}
	return string(b)
}

func GetFullURL(key string, m *model.StringMap) (val string, err error) {
	value, ok := (*m)[key]
	if !ok {
		return "", fmt.Errorf("id отсутствует")
	}
	return value, nil
}
