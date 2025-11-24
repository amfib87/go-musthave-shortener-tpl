package handler

import (
	"net/http"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
)

func IdGetHandler(res http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[1:]

	if id == "" {
		http.Error(res, "id is empty", http.StatusBadRequest)
		return
	}
	fullURL := service.GetFullURL(id)

	res.Header().Set("Location", fullURL)

	res.WriteHeader(http.StatusTemporaryRedirect)
}
