package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	gz "github.com/amfib87/go-musthave-shortener-tpl/internal/gzip"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
)

type Handler struct {
	cfg    *config.Cnfg
	mapURL *model.StringMap
	file   *os.File
	Logger *logger.TLog
}

func NewHandler(cfg *config.Cnfg, file *os.File, lg *logger.TLog) (h *Handler, err error) {
	data, err := service.InitMap(file)
	if err != nil {
		return nil, err
	}

	return &Handler{
		cfg:    cfg,
		mapURL: data,
		file:   file,
		Logger: lg,
	}, nil
}

func (h *Handler) PostURLHandler(res http.ResponseWriter, req *http.Request) {
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

	shortURL, err := service.GetShortURL(URL, h.mapURL, h.file)
	if err != nil {
		h.Logger.Lg.Sugar().Infoln("error GetShortURL: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	var serv string
	if h.cfg.AddrForURL == "" {
		val, err := url.JoinPath("http://", req.Host, "/", shortURL)
		if err != nil {
			h.Logger.Lg.Sugar().Infoln("failed to compose the shortened URL: %v", err)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		serv = val

	} else {
		val, err := url.JoinPath(h.cfg.AddrForURL, "/", shortURL)
		if err != nil {
			h.Logger.Lg.Sugar().Infoln("500 Internal Error: %v", err)
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

	fullURL, err := h.mapURL.GetFullURL(ID)
	if err != nil {
		h.Logger.Lg.Sugar().Infoln("500 Internal Error: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if fullURL == "" {
		h.Logger.Lg.Sugar().Infoln("id не найдено")
		http.Error(res, "id не найдено", http.StatusNotFound)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Location", fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) PostURLJSONHandler(res http.ResponseWriter, req *http.Request) {

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
		http.Error(res, "url is required", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &dataReq); err != nil {
		http.Error(res, "wrong JSON", http.StatusBadRequest)
		return
	}

	if dataReq.URL == "" {
		http.Error(res, "URL is empty", http.StatusBadRequest)
		return
	}

	dataAnsw.ShortURL, err = service.GetShortURL(dataReq.URL, h.mapURL, h.file)
	if err != nil {
		h.Logger.Lg.Sugar().Infoln("error GetShortURL: %v", err)
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
		h.Logger.Lg.Sugar().Infoln("failed to compose the shortened URL: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(dataAnsw)
	if err != nil {
		h.Logger.Lg.Sugar().Infoln("failed Marshal: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	res.Write(resp)

}

func (hndl *Handler) GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		origRes := res

		contentType := req.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			acceptEncoding := req.Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, "gzip") {
				newRes := gz.NewCompressWriter(res)
				origRes = newRes
				defer newRes.Close()
			}
		}

		contentEncoding := req.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			newReader, err := gz.NewCompressReader(req.Body)
			if err != nil {
				hndl.Logger.Lg.Sugar().Infoln("failed init NewCompressReader: %v", err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = newReader
			defer newReader.Close()
		}

		h.ServeHTTP(origRes, req)
	})
}
