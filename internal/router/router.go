package router

import (
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"
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
	r.chi.Get("/{id}", h.MainPostHandler)
	r.chi.Post("/", h.MainPostHandler)

	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
