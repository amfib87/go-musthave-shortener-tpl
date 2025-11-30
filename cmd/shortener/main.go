package main

import (
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/router"
)

func main() {

	err := http.ListenAndServe(":8080", router.Init())
	if err != nil {
		panic(err)
	}
}
