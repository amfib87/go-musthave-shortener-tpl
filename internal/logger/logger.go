package logger

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type TLog struct {
	Lg *zap.Logger
}

type (
	ResponseData struct {
		Status int
		Size   int
	}

	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) (lg *TLog, err error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed ParseAtomiclevel: %v", err)
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed cfg.Build: %v", err)
	}

	return &TLog{Lg: zl}, nil
}

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}

func (lg *TLog) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start := time.Now()

		responseData := &ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		// duration := time.Since(start)

		// lg.Lg.Sugar().Infoln(
		// 	"uri", r.RequestURI,
		// 	"method", r.Method,
		// 	"status", responseData.Status,
		// 	"duration", duration,
		// 	"size", responseData.Size,
		// )
	})
}
