package service

import (
	"fmt"
	"math/rand"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
)

var mapURL model.StringMap

func InitMap() {
	if mapURL == nil {
		mapURL = make(model.StringMap)
	}
}

func GetShortURL(key string) string {
	shortURL := generateShortID()
	mapURL[shortURL] = key
	return shortURL
}

func generateShortID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = model.Letters[rand.Intn(len(model.Letters))]
	}
	return string(b)
}

func GetFullURL(key string) (val string, err error) {
	value, ok := mapURL[key]
	if !ok {
		return "", fmt.Errorf("Id отсутствует")
	}
	return value, nil
}
