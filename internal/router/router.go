package router

import (
	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"
	"github.com/go-chi/chi"
)

func Init() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", handler.IDGetHandler)
	r.Post("/", handler.MainPostHandler)
	return r
}
