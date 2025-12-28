package router

import (
	"os"
	"testing"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	tempFile, err := os.CreateTemp(os.TempDir(), "Test9")
	if err != nil {
		t.Fatal("failed create temp file", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name()) // удаляем файл после теста
	}()

	// Используем имя созданного файла как StoragePath
	path := tempFile.Name()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg    *config.Cnfg
		router bool
		err    error
	}{
		{name: "Succs", cfg: &config.Cnfg{ServRunAddr: "", AddrForURL: "", StoragePath: path}, router: true, err: nil},
	}

	logger, err := logger.Initialize("Info")
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Init(tt.cfg, tempFile, logger)
			assert.NotNil(t, got, "Объект = nil")
			assert.Equal(t, err, tt.err)
		})
	}
}
