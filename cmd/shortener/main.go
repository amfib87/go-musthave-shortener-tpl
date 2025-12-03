package main

import (
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/router"
)

func main() {
	// обрабатываем аргументы командной строки
	config.ParseFlags()

	err := http.ListenAndServe(config.Cnfg.ServRunAddr, router.Init())
	if err != nil {
		panic(err)
	}
}
