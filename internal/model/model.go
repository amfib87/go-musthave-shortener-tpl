package model

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type TData map[string]string

var ErrKeyExists = errors.New("key already exists")

type StringMap struct {
	mu   sync.Mutex
	Data TData `json:"data"`
}

func (m *StringMap) InsertShortURL(key, shortURL string, f *os.File) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := (m.Data)[shortURL]; exists {
		return ErrKeyExists
	}

	(m.Data)[shortURL] = key
	if err := saveFile(m.Data, f); err != nil {
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

func saveFile(data TData, f *os.File) error {
	// сериализуем структуру
	dataJSON, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)

	// сохраняем данные в файл
	if _, err := writer.Write(dataJSON); err != nil {
		return err
	}
	writer.Flush()
	return nil
}
