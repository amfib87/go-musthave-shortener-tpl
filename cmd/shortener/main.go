package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/audit"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/router"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date:: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	const (
		certFile string = "server.crt"
		keyFile  string = "server.key"
	)

	logger, err := logger.Initialize("Info")
	if err != nil {
		log.Fatalf("failed init logger: %v", err)
		return err
	}

	// обрабатываем аргументы командной строки
	cfg := config.NewConfig()
	config.ParseFlags(cfg)
	logger.Lg.Info("cfg", zap.Any("cfg", cfg))

	urlStorage, err := service.InitURLStorage(cfg, logger)
	if err != nil {
		logger.Lg.Error("failed IniturlStorage:", zap.Error(err))
		return err
	}
	defer urlStorage.Close(logger)

	// Инициализируем аудит
	audit, err := audit.NewAuditManager(cfg.AuditFile, cfg.AddrForURL)
	if err != nil {
		logger.Lg.Error("failed audit.InitAudit", zap.Error(err))
		return err
	}

	// Инициализируем маршрутизатор с конфигурацией
	router, err := router.Init(cfg, logger, urlStorage, audit)
	if err != nil {
		logger.Lg.Sugar().Fatalf("failed to init router: %v", err)
		return err
	}

	tlsConfig := &tls.Config{}
	if cfg.EnablHttps != "" {

		// Проверяем существование файлов сертификата и ключа
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			logger.Lg.Sugar().Infoln("Certificate not found, generating new self‑signed certificate...")
			err := service.GenerateTLSCertificate(certFile, keyFile)
			if err != nil {
				logger.Lg.Sugar().Fatalf("Failed to generate certificate: ", err)
			}
		}

		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			logger.Lg.Sugar().Infoln("Private key not found, generating new private key...")
			err := service.GenerateTLSCertificate(certFile, keyFile)
			if err != nil {
				logger.Lg.Sugar().Fatalf("Failed to generate key: ", err)
			}
		}

		tlsConfig.MinVersion = tls.VersionTLS12
		tlsConfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		}
	}

	server := &http.Server{
		Addr:    cfg.ServRunAddr,
		Handler: router,
	}

	if cfg.EnablHttps != "" {
		server.TLSConfig = tlsConfig
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if cfg.EnablHttps != "" {

			logger.Lg.Info("Running HTTPS server", zap.String("address", cfg.ServRunAddr))
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				logger.Lg.Fatal("Server failed: %v", zap.Error(err))
			}

		} else {

			logger.Lg.Info("Running server", zap.String("address", cfg.ServRunAddr))
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Lg.Fatal("Server failed: %v", zap.Error(err))
			}
		}
	}()

	<-stop
	logger.Lg.Info("Shutdown signal received")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Lg.Error("Graceful shutdown failed: %v", zap.Error(err))
		// Принудительное закрытие
		server.Close()
	}
	logger.Lg.Info("Server stopped gracefully")

	return nil
}
