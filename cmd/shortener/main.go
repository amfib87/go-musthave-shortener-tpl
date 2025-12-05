package main

import (
	"log"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/router"
)

func main() {
	// обрабатываем аргументы командной строки
	cfg := config.NewConfig()
	config.ParseFlags(cfg)

	// Инициализируем маршрутизатор с конфигурацией
	router := router.Init(cfg)
	log.Fatal(http.ListenAndServe(cfg.ServRunAddr, router))
}
