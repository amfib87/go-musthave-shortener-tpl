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

	if envServAddr, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		cfg.ServRunAddr = envServAddr
	}
	if envAddrForURL, exist := os.LookupEnv("BASE_URL"); exist {
		cfg.AddrForURL = envAddrForURL
	}

	if envStoragePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		cfg.StoragePath = envStoragePath
	}

	if cfg.StoragePath == "" {
		path, err := os.UserHomeDir()
		if err != nil || path == "" {
			path = "/var/lib/myapp"
		}

		cfg.StoragePath = filepath.Join(path, "Iter9")
	}

	if envDataBaseDsn, exist := os.LookupEnv("DATABASE_DSN"); exist {
		cfg.DataBaseDsn = envDataBaseDsn
	}
}
