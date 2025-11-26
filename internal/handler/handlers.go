package handler

import (
	"io"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
)

func MainPostHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Тип запроса некорректный", http.StatusMethodNotAllowed)
		return
	}

	service.InitMap()

	body, err := io.ReadAll(req.Body)
	URL := string(body)

	if err != nil || URL == "" {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := service.GetShortURL(URL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://localhost:8080/" + shortURL))
}

func IDGetHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Тип запроса некорректный", http.StatusMethodNotAllowed)
		return
	}

	if req.URL.Path == "" {
		http.Error(res, "ID is empty", http.StatusBadRequest)
		return
	}
	ID := req.URL.Path[1:]
	fullURL := service.GetFullURL(ID)

	if fullURL == "" {
		http.Error(res, "URL не найден", http.StatusNotFound)
		return
	}

	res.Header().Set("Location", fullURL)

	res.WriteHeader(http.StatusTemporaryRedirect)
}
