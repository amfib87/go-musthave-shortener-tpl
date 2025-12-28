package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/go-chi/chi"
)

type Router struct {
	chi *chi.Mux
}

func Init(cfg *config.Cnfg, file *os.File, lg *logger.TLog) (*Router, error) {
	r := &Router{
		chi: chi.NewRouter(),
	}

	// Создаём обработчик с конфигурацией
	h, err := handler.NewHandler(cfg, file, lg)
	if err != nil {
		return nil, fmt.Errorf("failed NewHandler: %v", err)
	}

	//Middlieware
	r.chi.Use(h.Logger.RequestLogger)
	r.chi.Use(h.GzipMiddleware)

	// Регистрируем маршруты
	r.chi.Get("/{id}", h.IDGetHandler)
	r.chi.Post("/", h.PostURLHandler)
	r.chi.Post("/{api}/{shorten}", h.PostURLJSONHandler)
	return r, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
