package main

import (
	"log"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/router"
)

func main() {
	// обрабатываем аргументы командной строки
	cfg := config.NewConfig()
	config.ParseFlags(cfg)

	// Инициализируем маршрутизатор с конфигурацией
	router, err := router.Init(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := logger.Initialize("Info"); err != nil {
		log.Fatal(err.Error())
	}

	log.Fatal(http.ListenAndServe(cfg.ServRunAddr, router))
}
