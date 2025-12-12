package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		filled bool
	}{
		// TODO: Add test cases.
		{name: "config not nill", filled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig()
			if tt.filled {
				assert.NotNil(t, got, "failed init config")
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name           string // description of this test case
		envVars        map[string]string
		args           []string
		expServRunAddr string
		expAddrForURL  string
	}{
		{
			name:           "no flags, no env vars",
			envVars:        map[string]string{},
			args:           []string{},
			expServRunAddr: ":8080",
			expAddrForURL:  "",
		},
		{
			name:           "flags only",
			envVars:        map[string]string{},
			args:           []string{"-a", "localhost:8080", "-b", "/HHHttt"},
			expServRunAddr: "localhost:8080",
			expAddrForURL:  "/HHHttt",
		},
		{
			name: "env vars only",
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:9000",
				"BASE_URL":       "/KKKooo",
			},
			args:           []string{},
			expServRunAddr: "localhost:9000",
			expAddrForURL:  "/KKKooo",
		},
		{
			name: "flags and env vars (flags should override)",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env-localhost:8080",
				"BASE_URL":       "/YYYYuuuu",
			},
			args:           []string{"-a", "flag-localhost:9000", "-b", "/ffffffff"},
			expServRunAddr: "env-localhost:8080",
			expAddrForURL:  "/YYYYuuuu",
		},
		{
			name: "only SERVER_ADDRESS env var",
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:7000",
			},
			args:           []string{},
			expServRunAddr: "localhost:7000",
			expAddrForURL:  "",
		},
		{
			name: "only BASE_URL env var",
			envVars: map[string]string{
				"BASE_URL": "/55555555",
			},
			args:           []string{},
			expServRunAddr: ":8080",
			expAddrForURL:  "/55555555",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			// Устанавливаем переменные окружения
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Имитируем аргументы командной строки
			os.Args = append([]string{"cmd"}, tt.args...)

			// Вызываем тестируемую функцию
			cfg := &Cnfg{}
			ParseFlags(cfg)

			// Проверяем результаты
			if cfg.ServRunAddr != tt.expServRunAddr {
				t.Errorf("ParseFlags() cfg.ServRunAddr = %q, want %q", cfg.ServRunAddr, tt.expServRunAddr)
			}
			if cfg.AddrForURL != tt.expAddrForURL {
				t.Errorf("ParseFlags() cfg.AddrForURL = %q, want %q", cfg.AddrForURL, tt.expAddrForURL)
			}
		})
	}
}
