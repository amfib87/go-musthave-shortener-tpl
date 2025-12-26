package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type TData map[string]string

var ErrKeyExists = errors.New("key already exists")

type StringMap struct {
	Mu   sync.Mutex
	Data TData `json:"data"`
	Name string
}

func (m *StringMap) InsertShortURL(key, shortURL string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := (m.Data)[shortURL]; exists {
		return ErrKeyExists
	}

	(m.Data)[shortURL] = key
	if err := saveFile(m.Data, m.Name); err != nil {
		return err
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

func saveFile(data TData, name string) error {
	// сериализуем структуру в JSON формат
	dataJSON, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return err
	}
	// сохраняем данные в файл
	if err := os.WriteFile(name, dataJSON, 0666); err != nil {
		return err
	}
	return nil
}
