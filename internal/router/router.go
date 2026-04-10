// Package router предназначен для работы с маршрутизатором функций-обработчиков
package router

import (
	"fmt"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/audit"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	"github.com/go-chi/chi"
)

type Router struct {
	chi *chi.Mux
}

func Init(cfg *config.Cnfg, lg *logger.TLog, st service.URLStorage, au *audit.AuditManager) (*Router, error) {
	r := &Router{
		chi: chi.NewRouter(),
	}

	// Создаём обработчик
	h, err := handler.NewHandler(cfg, lg, st, au)
	if err != nil {
		return nil, fmt.Errorf("failed NewHandler: %v", err)
	}

	//Middlieware
	r.chi.Use(h.Logger.RequestLogger)
	r.chi.Use(h.GzipMiddleware)
	r.chi.Use(h.TimeoutMiddleware)
	r.chi.Use(h.AuthCookieMiddleware)

	// Регистрируем маршруты
	r.chi.Get("/{id}", h.IDGetHandler)
	r.chi.Get("/ping", h.GetPing)
	r.chi.Get("/{api}/{user}/{urls}", h.GetAllURLsHandler)

	r.chi.Post("/", h.PostURLHandler)
	r.chi.Post("/{api}/{shorten}", h.PostURLJSONHandler)
	r.chi.Post("/{api}/{shorten}/{batch}", h.PostMassURLHandler)

	r.chi.Delete("/{api}/{user}/{urls}", h.DelShortURLsHandler)
	return r, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
