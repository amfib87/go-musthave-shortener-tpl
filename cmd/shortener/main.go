package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/repository"
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

// func run() error {
// 	logger, err := logger.Initialize("Info")
// 	if err != nil {
// 		log.Fatalf("failed init logger: %v", err)
// 		return err
// 	}

// 	// обрабатываем аргументы командной строки
// 	cfg := config.NewConfig()
// 	config.ParseFlags(cfg)

// 	var db *sql.DB
// 	var file *os.File

// 	if cfg.DataBaseDsn != "" {
// 		db, err = repository.InitDB(cfg.DataBaseDsn)
// 		if err != nil {
// 			logger.Lg.Sugar().Fatalf("failed InitDB: %v", err)
// 			return err
// 		}
// 		defer db.Close()

// 	} else if cfg.StoragePath != "" {
// 		file, err = service.InitFile(cfg.StoragePath)
// 		if err != nil {
// 			logger.Lg.Sugar().Fatalf("failed to init file: %v", err)
// 			return err
// 		}
// 		defer service.FileClose(file, logger)
// 	}

// 	// Инициализируем маршрутизатор с конфигурацией
// 	router, err := router.Init(cfg, file, logger, db)
// 	if err != nil {
// 		logger.Lg.Sugar().Fatalf("failed to init router: %v", err)
// 		return err
// 	}

// 	// logger.Lg.Sugar().Fatalf("failed listenServer: %v", http.ListenAndServe(cfg.ServRunAddr, router))

// 	logger.Lg.Info("Running server", zap.String("address", cfg.ServRunAddr))
// 	return http.ListenAndServe(cfg.ServRunAddr, router)
// }

func run() error {
	logger, err := logger.Initialize("Info")
	if err != nil {
		log.Fatalf("failed init logger: %v", err)
		return err
	}

	cfg := config.NewConfig()
	config.ParseFlags(cfg)

	var db *sql.DB
	var file *os.File

	if cfg.DataBaseDsn != "" {
		db, err = repository.InitDB(cfg.DataBaseDsn)
		if err != nil {
			logger.Lg.Sugar().Fatalf("failed InitDB: %v", err)
			return err
		}
		defer db.Close()
	} else if cfg.StoragePath != "" {
		file, err = service.InitFile(cfg.StoragePath)
		if err != nil {
			logger.Lg.Sugar().Fatalf("failed to init file: %v", err)
			return err
		}
		defer service.FileClose(file, logger)
	}

	// Инициализация сервера (без запуска)
	server, err := StartServer(cfg, file, logger, db)
	if err != nil {
		logger.Lg.Sugar().Fatal("failed to init server: ", err)
		return err
	}

	// Обработчик сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Lg.Info("Получен сигнал остановки. Завершаем работу...")

		if err := StopServer(server, 3*time.Second); err != nil {
			logger.Lg.Error("Ошибка при остановке сервера", zap.Error(err))
		}
	}()

	logger.Lg.Info("Запуск сервера", zap.String("адрес", cfg.ServRunAddr))

	// Запуск сервера
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func StartServer(cfg *config.Cnfg, file *os.File, logger *logger.TLog, db *sql.DB) (*http.Server, error) {
	router, err := router.Init(cfg, file, logger, db)
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:         cfg.ServRunAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server, nil
}

// StopServer корректно останавливает сервер
func StopServer(server *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed stop server: %w", err)
	}
	return nil
}
