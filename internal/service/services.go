package service

import (
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
	_, ok := mapURL[key]
	if !ok {
		shortURL := generateShortId()
		mapURL[key] = shortURL
		return shortURL
	}
	return ""
}

func generateShortId() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = model.Letters[rand.Intn(len(model.Letters))]
	}
	return string(b)
}

func GetFullURL(id string) string {
	for key, val := range mapURL {
		if val == id {
			return key
		}
	}
	return ""
}
