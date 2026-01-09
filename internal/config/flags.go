package config

import (
	"flag"
	"os"
	"path/filepath"
)

type Cnfg struct {
	ServRunAddr string
	AddrForURL  string
	StoragePath string
}

func NewConfig() *Cnfg {
	return &Cnfg{
		ServRunAddr: "",
		AddrForURL:  "",
		StoragePath: "",
	}
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags(cfg *Cnfg) {
	flag.StringVar(&cfg.ServRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.AddrForURL, "b", "", "address to short URL")
	flag.StringVar(&cfg.StoragePath, "f", "", "path file for storage")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if envServAddr, _ := os.LookupEnv("SERVER_ADDRESS"); envServAddr != "" {
		cfg.ServRunAddr = envServAddr
	}
	if envAddrForURL, _ := os.LookupEnv("BASE_URL"); envAddrForURL != "" {
		cfg.AddrForURL = envAddrForURL
	}

	if envStoragePath, _ := os.LookupEnv("FILE_STORAGE_PATH"); envStoragePath != "" {
		cfg.StoragePath = envStoragePath
	}

	if cfg.StoragePath == "" {
		path, err := os.UserHomeDir()
		if err != nil || path == "" {
			path = "/var/lib/myapp"
		}

		cfg.StoragePath = filepath.Join(path, "Iter9")
	}
}
