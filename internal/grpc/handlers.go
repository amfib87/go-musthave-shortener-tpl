package grpc

import (
	"context"
	"errors"
	"net/url"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	pb "github.com/amfib87/go-musthave-shortener-tpl/proto"
)

func (s *Server) ListUserURLs(ctx context.Context, req *pb.ListUserURLsRequest) (*pb.UserURLsResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, errors.New("user ID not found in context")
	}

	allURLs := s.Handler.MapURL.GetAllURLsForUser(userID)

	baseURL := s.Handler.Cfg.AddrForURL
	if baseURL == "" {
		baseURL = "http://" + s.Handler.Cfg.ServRunAddr
	}

	builder := NewUserURLsResponseBuilder()
	for short, full := range allURLs {
		shortExp, err := url.JoinPath(baseURL, "/", short)
		if err != nil {
			continue
		}

		builder.AddURL(&pb.URLData{ShortUrl: shortExp,
			OriginalUrl: full})
	}

	return builder.Build(), nil

}

func (s *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	// originalURL := req.GetUrl()
	// shortURL := s.generateShortURL(originalURL)

	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, errors.New("user ID not found in context")
	}

	dataRow := model.DataRow{
		URL:    req.Url,
		UserID: userID,
	}

	shortURL, err := service.GetShortURL(ctx, dataRow, s.Handler.MapURL, s.Handler.URLSt, s.Handler.Logger)
	if err != nil {
		return nil, err
	}

	baseURL := s.Handler.Cfg.AddrForURL
	if baseURL == "" {
		baseURL = "http://" + s.Handler.Cfg.ServRunAddr
	}

	fullShortURL, err := url.JoinPath(baseURL, "/", shortURL)
	if err != nil {
		return nil, err
	}

	response := NewURLShortenResponseBuilder().
		SetResult(fullShortURL).
		Build()

	return response, nil
}

func (s *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	dataRow, err := s.Handler.MapURL.GetFullURL(req.Id)
	if err != nil || dataRow.URL == "" {
		return nil, errors.New("URL not found")
	}

	response := &pb.URLExpandResponse{
		Result: dataRow.URL,
	}

	return response, nil
}
