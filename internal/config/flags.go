package config

import (
	"flag"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
var Cnfg struct {
	ServRunAddr string
	AddrForURL  string
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&Cnfg.ServRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&Cnfg.AddrForURL, "b", "", "address to short URL")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
