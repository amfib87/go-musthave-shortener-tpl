package router

import (
	"os"
	"testing"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/audit"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
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

	path := tempFile.Name()

	tests := []struct {
		name   string // description of this test case
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
			got, err := Init(tt.cfg, logger, service.URLStorage{}, &audit.AuditManager{})
			assert.NotNil(t, got, "Объект = nil")
			assert.Equal(t, err, tt.err)
		})
	}
}
