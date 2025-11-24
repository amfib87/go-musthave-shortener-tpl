package main

import (
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handler.MainPostHandler)
	mux.HandleFunc("GET /", handler.IDGetHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
