package handler

import (
	"io"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
)

func MainPostHandler(res http.ResponseWriter, req *http.Request) {
	service.InitMap()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte(err.Error()))
		return
	}

	url := string(body)
	shortURL := service.GetShortURL(url)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://localhost:8080/" + shortURL))
}
