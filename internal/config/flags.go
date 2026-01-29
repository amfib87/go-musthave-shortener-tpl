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
	DataBaseDsn string
}

func NewConfig() *Cnfg {
	return &Cnfg{
		ServRunAddr: "",
		AddrForURL:  "",
		StoragePath: "",
		DataBaseDsn: "",
	}
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags(cfg *Cnfg) {
	flag.StringVar(&cfg.ServRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.AddrForURL, "b", "", "address to short URL")
	flag.StringVar(&cfg.StoragePath, "f", "", "path file for storage")
	flag.StringVar(&cfg.DataBaseDsn, "d", "", "address BD")
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
}
