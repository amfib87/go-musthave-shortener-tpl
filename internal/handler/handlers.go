package handler

import (
	"bytes"
	"encoding/json"
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

func NewHandler(cfg *config.Cnfg) (h *Handler, err error) {
	data, err := service.InitMap(cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	return &Handler{
		cfg:    cfg,
		mapURL: data,
	}, nil
}

func (h *Handler) MainPostHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	URL := string(body)
	if URL == "" {
		http.Error(res, "url is empty", http.StatusBadRequest)
		return
	}

	shortURL, err := service.GetShortURL(URL, h.mapURL)
	if err != nil {
		log.Printf("error GetShortURL: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	var serv string
	if h.cfg.AddrForURL == "" {
		val, err := url.JoinPath("http://", req.Host, "/", shortURL)
		if err != nil {
			log.Printf("failed to compose the shortened URL: %v", err)
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
		log.Println("ID is empty")
		http.Error(res, "id is required", http.StatusBadRequest)
		return
	}

	fullURL, err := h.mapURL.GetFullURL(ID)
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

func (h *Handler) PostShortenHandler(res http.ResponseWriter, req *http.Request) {

	type dataRequest struct {
		URL string `json:"url"`
	}

	type dataAnswer struct {
		ShortURL string `json:"result"`
	}

	var dataReq dataRequest
	var dataAnsw dataAnswer
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		log.Printf("failed read body of request: %v", err)
		http.Error(res, "url is required", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &dataReq); err != nil {
		log.Printf("failed unmarshal: %v", err)
		http.Error(res, "wrong JSON", http.StatusBadRequest)
		return
	}

	if dataReq.URL == "" {
		log.Printf("URL is empty")
		http.Error(res, "URL is empty", http.StatusBadRequest)
		return
	}

	dataAnsw.ShortURL, err = service.GetShortURL(dataReq.URL, h.mapURL)
	if err != nil {
		log.Printf("error GetShortURL: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	baseURL := h.cfg.AddrForURL
	if baseURL == "" {
		baseURL = "http://" + req.Host
	}

	dataAnsw.ShortURL, err = url.JoinPath(baseURL, "/", dataAnsw.ShortURL)
	if err != nil {
		log.Printf("failed to compose the shortened URL: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(dataAnsw)
	if err != nil {
		log.Printf("failed Marshal: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	res.Write(resp)

}
