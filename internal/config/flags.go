// Package config для получения аргументов командной строки и значения переменных окружения
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
)

type Cnfg struct {
	ServRunAddr string `json:"server_address"`
	AddrForURL  string `json:"base_url"`
	StoragePath string `json:"file_storage_path"`
	DataBaseDsn string `json:"database_dsn"`
	AuditFile   string
	AuditURL    string
	EnablHttps  bool `json:"enable_https"`
}

func NewConfig() *Cnfg {
	return &Cnfg{
		ServRunAddr: "",
		AddrForURL:  "",
		StoragePath: "",
		DataBaseDsn: "",
	}
}

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags(cfg *Cnfg, log *logger.TLog) {
	flag.StringVar(&cfg.ServRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.AddrForURL, "b", "", "address to short URL")
	flag.StringVar(&cfg.StoragePath, "f", "", "path file for storage")
	flag.StringVar(&cfg.DataBaseDsn, "d", "", "address BD")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "address for audit file")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "URL for audit")
	flag.BoolVar(&cfg.EnablHttps, "s", false, "Enable HTTPS server")

	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if envServAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServRunAddr = envServAddr
	}
	if envAddrForURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.AddrForURL = envAddrForURL
	}

	if envStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.StoragePath = envStoragePath
	}

	if cfg.StoragePath == "" {
		path, err := os.UserHomeDir()
		if err != nil || path == "" {
			path = "/var/lib/myapp"
		}

		cfg.StoragePath = filepath.Join(path, "Iter9")
	}

	if envDataBaseDsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DataBaseDsn = envDataBaseDsn
	}

	if envAuditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL, ok := os.LookupEnv("AUDIT_URL"); ok {
		cfg.AuditURL = envAuditURL
	}

	if envEnablHttps, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		EnablHttpsBool, err := strconv.ParseBool(envEnablHttps)
		if err != nil {
			log.Lg.Sugar().Errorln("failed get strconv.ParseBool", err)
		}
		cfg.EnablHttps = EnablHttpsBool
	}

	var configFile string
	flag.StringVar(&configFile, "c", "", "config file path")
	flag.StringVar(&configFile, "config", "", "config file path")

	if err := readConfigFile(configFile, cfg, log); err != nil {
		log.Lg.Sugar().Errorln("Failed to load configuration: %v", err)
	}
}

func readConfigFile(configFile string, cfg *Cnfg, log *logger.TLog) error {
	if configFile == "" {
		return nil // Файл не указан — пропускаем
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Lg.Sugar().Errorln("Config file %s not found, skipping", configFile)
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var fileConfig Cnfg
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Применяем только те поля, которые ещё не заданы флагами/окружением
	if cfg.ServRunAddr == "" {
		cfg.ServRunAddr = fileConfig.ServRunAddr
	}
	if cfg.AddrForURL == "" {
		cfg.AddrForURL = fileConfig.AddrForURL
	}
	if cfg.StoragePath == "" {
		cfg.StoragePath = fileConfig.StoragePath
	}
	if cfg.DataBaseDsn == "" {
		cfg.DataBaseDsn = fileConfig.DataBaseDsn
	}
	if !cfg.EnablHttps { // Флаг для отслеживания, было ли значение задано извне
		cfg.EnablHttps = fileConfig.EnablHttps
	}

	return nil
}
