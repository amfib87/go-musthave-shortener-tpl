package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	gz "github.com/amfib87/go-musthave-shortener-tpl/internal/gzip"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	cfg    *config.Cnfg
	mapURL *model.StringMap
	Logger *logger.TLog
	urlSt  service.URLStorage
}

func NewHandler(cfg *config.Cnfg, lg *logger.TLog, st service.URLStorage) (h *Handler, err error) {
	data, err := service.InitMap(st)
	if err != nil {
		return nil, err
	}

	return &Handler{
		cfg:    cfg,
		mapURL: data,
		Logger: lg,
		urlSt:  st,
	}, nil
}

func (hndl *Handler) PostURLHandler(res http.ResponseWriter, req *http.Request) {
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

	shortURL, err := service.GetShortURL(req.Context(), URL, hndl.mapURL, hndl.urlSt, hndl.Logger)
	if err == model.ErrOriginalURLExist {
		hndl.Logger.Lg.Sugar().Infoln("error GetShortURL: %v", err.Error())

		val, err := url.JoinPath("http://", req.Host, "/", shortURL)
		if err != nil {
			hndl.Logger.Lg.Error("failed to compose the shortened URL:", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusConflict)
		res.Write([]byte(val))
		return
	}

	if err != nil {
		hndl.Logger.Lg.Error("error GetShortURL:", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	var serv string
	if hndl.cfg.AddrForURL == "" {
		val, err := url.JoinPath("http://", req.Host, "/", shortURL)
		if err != nil {
			hndl.Logger.Lg.Error("failed to compose the shortened URL:", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		serv = val

	} else {
		val, err := url.JoinPath(hndl.cfg.AddrForURL, "/", shortURL)
		if err != nil {
			hndl.Logger.Lg.Error("500 Internal Error: %v", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		serv = val
	}

	res.Write([]byte(serv))
}

func (hndl *Handler) IDGetHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "" {
		http.Error(res, "id is empty", http.StatusBadRequest)
		return
	}

	ID := req.URL.Path[1:]
	if ID == "" {
		http.Error(res, "id is required", http.StatusBadRequest)
		return
	}

	fullURL, err := hndl.mapURL.GetFullURL(ID)
	if err != nil {
		hndl.Logger.Lg.Error("500 Internal Error: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if fullURL == "" {
		hndl.Logger.Lg.Sugar().Infoln("id не найдено")
		http.Error(res, "id не найдено", http.StatusNotFound)
		return
	}

	res.Header().Set("Location", fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (hndl *Handler) PostURLJSONHandler(res http.ResponseWriter, req *http.Request) {

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

	dataAnsw.ShortURL, err = service.GetShortURL(req.Context(), dataReq.URL, hndl.mapURL, hndl.urlSt, hndl.Logger)

	if errors.Is(err, model.ErrOriginalURLExist) {
		hndl.Logger.Lg.Sugar().Debugln("error GetShortURL: %v", err.Error())

		dataAnsw.ShortURL, err = url.JoinPath("http://", req.Host, "/", dataAnsw.ShortURL)
		if err != nil {
			hndl.Logger.Lg.Error("failed to compose the shortened URL", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(dataAnsw)
		if err != nil {
			hndl.Logger.Lg.Error("failed Marshal:", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusConflict)
		res.Write([]byte(resp))
		return
	}

	if err != nil {
		hndl.Logger.Lg.Error("error GetShortURL:", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	baseURL := hndl.cfg.AddrForURL
	if baseURL == "" {
		baseURL = "http://" + req.Host
	}

	dataAnsw.ShortURL, err = url.JoinPath(baseURL, "/", dataAnsw.ShortURL)
	if err != nil {
		hndl.Logger.Lg.Error("failed to compose the shortened URL:", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(dataAnsw)
	if err != nil {
		hndl.Logger.Lg.Error("failed Marshal:", zap.Error(err))
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
				hndl.Logger.Lg.Error("failed init NewCompressReader:", zap.Error(err))
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = newReader
			defer newReader.Close()
		}

		h.ServeHTTP(origRes, req)
	})
}

func (hndl *Handler) TimeoutMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		resTimeout := http.TimeoutHandler(h, 12*time.Second, "Timeout expired: The server did not respond in time")

		resTimeout.ServeHTTP(res, req)
	})
}

func (hndl *Handler) GetPing(res http.ResponseWriter, req *http.Request) {
	if er := hndl.urlSt.DB.PingContext(req.Context()); er != nil {
		hndl.Logger.Lg.Error("failed PingContext", zap.Error(er))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(""))
}

func (hndl *Handler) PostMassURLHandler(res http.ResponseWriter, req *http.Request) {

	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, "url is required", http.StatusBadRequest)
		return
	}

	var dataReq []model.DataRequestMass
	if err = json.Unmarshal(buf.Bytes(), &dataReq); err != nil {
		http.Error(res, "wrong JSON", http.StatusBadRequest)
		return
	}

	dataAnsw, err := service.GetShortURLMass(req.Context(), dataReq, hndl.mapURL, hndl.urlSt)
	if err != nil {
		hndl.Logger.Lg.Error("error GetShortURL:", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	baseURL := hndl.cfg.AddrForURL
	if baseURL == "" {
		baseURL = "http://" + req.Host
	}

	for ind, lineAnswer := range dataAnsw {
		lineAnswer.ShortURL, err = url.JoinPath(baseURL, "/", lineAnswer.ShortURL)
		if err != nil {
			hndl.Logger.Lg.Error("failed to compose the shortened URL: %v", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		dataAnsw[ind] = lineAnswer
	}

	resp, err := json.Marshal(dataAnsw)
	if err != nil {
		hndl.Logger.Lg.Error("failed Marshal: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	res.Write(resp)

}
