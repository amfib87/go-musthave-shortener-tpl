package model

import (
	"errors"
	"fmt"
	"sync"
)

var ErrKeyExists = errors.New("key already exists")

type StringMap struct {
	Mu   sync.Mutex
	Data map[string]string
}

func (m *StringMap) InsertShortURL(key, shortURL string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := (m.Data)[shortURL]; exists {
		return ErrKeyExists
	}

	(m.Data)[shortURL] = key
	return nil
}

func (m *StringMap) GetFullURL(key string) (val string, err error) {
	value, ok := (m.Data)[key]
	if !ok {
		return "", fmt.Errorf("id отсутствует")
	}
	return value, nil
}
