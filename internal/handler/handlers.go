package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
)

type Handler struct {
	cfg    *config.Cnfg
	mapURL *model.StringMap
}

func NewHandler(cfg *config.Cnfg) *Handler {
	return &Handler{
		cfg:    cfg,
		mapURL: service.InitMap(),
	}
}

func (h *Handler) MainPostHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	URL := string(body)
	if URL == "" {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := service.GetShortURL(URL, h.mapURL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	var serv string
	if h.cfg.AddrForURL == "" {
		val, err := url.JoinPath("http://", req.Host, "/", shortURL)
		if err != nil {
			log.Printf("500 Internal Error: %v", err)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		serv = val

	} else {
		val, err := url.JoinPath(h.cfg.AddrForURL, "/", shortURL)
		if err != nil {
			log.Printf("500 Internal Error: %v", err)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		serv = val
	}

	res.Write([]byte(serv))
}

func (h *Handler) IDGetHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "" {
		http.Error(res, "id is empty", http.StatusBadRequest)
		return
	}
	ID := req.URL.Path[1:]
	if ID == "" {
		http.Error(res, "id is required", http.StatusBadRequest)
		return
	}

	fullURL, err := service.GetFullURL(ID, h.mapURL)
	if err != nil {
		log.Printf("500 Internal Error: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
