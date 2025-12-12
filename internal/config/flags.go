package config

import (
	"flag"
	"os"
)

type Cnfg struct {
	ServRunAddr string
	AddrForURL  string
}

func NewConfig() *Cnfg {
	return &Cnfg{
		ServRunAddr: "",
		AddrForURL:  "",
	}
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags(cfg *Cnfg) {
	flag.StringVar(&cfg.ServRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.AddrForURL, "b", "", "address to short URL")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if envServAddr := os.Getenv("SERVER_ADDRESS"); envServAddr != "" {
		cfg.ServRunAddr = envServAddr
	}
	if envAddrForURL := os.Getenv("BASE_URL"); envAddrForURL != "" {
		cfg.AddrForURL = envAddrForURL
	}
}
