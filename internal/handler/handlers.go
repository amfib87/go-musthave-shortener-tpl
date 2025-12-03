package handler

import (
	"io"
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
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
	// res.Write([]byte("http://localhost:8080/" + shortURL))

	var serv string
	if config.Cnfg.AddrForURL == "" {
		serv = "http://" + req.Host + "/" + shortURL
	} else {
		serv = config.Cnfg.AddrForURL + "/" + shortURL
	}
	res.Write([]byte(serv))
}

func IDGetHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "тип запроса некорректный", http.StatusMethodNotAllowed)
		return
	}

	if req.URL.Path == "" {
		http.Error(res, "id is empty", http.StatusBadRequest)
		return
	}
	ID := req.URL.Path[1:]
	if ID == "" {
		http.Error(res, "id is required", http.StatusBadRequest)
		return
	}

	fullURL, err := service.GetFullURL(ID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}
	if fullURL == "" {
		http.Error(res, "id не найдено", http.StatusNotFound)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Location", fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
