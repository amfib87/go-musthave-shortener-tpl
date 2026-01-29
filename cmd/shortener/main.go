package main

import (
	"log"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/router"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	logger, err := logger.Initialize("Info")
	if err != nil {
		log.Fatalf("failed init logger: %v", err)
		return err
	}

	// обрабатываем аргументы командной строки
	cfg := config.NewConfig()
	config.ParseFlags(cfg)

	urlStorage, err := service.InitURLStorage(cfg, logger)
	if err != nil {
		logger.Lg.Error("failed IniturlStorage: %s", zap.Error(err))
		return err
	}
	defer urlStorage.Close(logger)

	// Инициализируем маршрутизатор с конфигурацией
	router, err := router.Init(cfg, logger, urlStorage)
	if err != nil {
		logger.Lg.Sugar().Fatalf("failed to init router: %v", err)
		return err
	}

	logger.Lg.Info("Running server", zap.String("address", cfg.ServRunAddr))
	err = http.ListenAndServe(cfg.ServRunAddr, router)
	logger.Lg.Sugar().Fatalf("failed listenServer: %v", err)
	return err
}
