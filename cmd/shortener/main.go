package main

import (
	"log"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/router"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger, err := logger.Initialize("Info")
	if err != nil {
		log.Fatalf("failed init logger: %v", err)
	}

	// обрабатываем аргументы командной строки
	cfg := config.NewConfig()
	config.ParseFlags(cfg)

	db, err := store.InitDB(cfg.DataBaseDsn)
	if err != nil {
		logger.Lg.Sugar().Fatalf("failed InitDB: %v", err)
	}
	defer db.Close()

	file, err := service.InitFile(cfg.StoragePath)
	if err != nil {
		logger.Lg.Sugar().Fatalf("failed to init file: %v", err)
	}
	defer service.FileClose(file, logger)

	// Инициализируем маршрутизатор с конфигурацией
	router, err := router.Init(cfg, file, logger, db)
	if err != nil {
		logger.Lg.Sugar().Fatalf("failed to init router: %v", err)
	}

	logger.Lg.Sugar().Fatalf("failed listenServer: %v", http.ListenAndServe(cfg.ServRunAddr, router))
}
