package router

import (
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"
	log "github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/go-chi/chi"
)

type Router struct {
	chi *chi.Mux
}

// Init создаёт и настраивает маршрутизатор с конфигурацией
func Init(cfg *config.Cnfg) *Router {
	r := &Router{
		chi: chi.NewRouter(),
	}

	// Создаём обработчик с конфигурацией
	h := handler.NewHandler(cfg)

	// Регистрируем маршруты
	r.chi.Get("/{id}", log.RequestLogger(h.IDGetHandler))
	r.chi.Post("/", log.RequestLogger(h.MainPostHandler))
	r.chi.Post("/{api}/{shorten}", log.RequestLogger(h.PostShortenHandler))
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
