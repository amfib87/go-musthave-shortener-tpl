package router

import (
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/gzip"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"
	log "github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/go-chi/chi"
)

type Router struct {
	chi *chi.Mux
}

// Init создаёт и настраивает маршрутизатор с конфигурацией
func Init(cfg *config.Cnfg) (*Router, error) {
	r := &Router{
		chi: chi.NewRouter(),
	}

	// Создаём обработчик с конфигурацией
	h, err := handler.NewHandler(cfg)
	if err != nil {
		return nil, err
	}

	// Регистрируем маршруты
	r.chi.Get("/{id}", log.RequestLogger(gzip.GzipMiddleware(h.IDGetHandler)))
	r.chi.Post("/", log.RequestLogger(gzip.GzipMiddleware(h.MainPostHandler)))
	r.chi.Post("/{api}/{shorten}", log.RequestLogger(gzip.GzipMiddleware(h.PostShortenHandler)))
	return r, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
