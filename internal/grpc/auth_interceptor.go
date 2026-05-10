package grpc

import (
	"context"
	"errors"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	userIDKey           = "userID"
)

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}

	authHeader, ok := md[authorizationHeader]
	if !ok || len(authHeader) == 0 {
		return nil, errors.New("authorization header required")
	}

	userID, err := service.GetUserID(authHeader[0])
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, userIDKey, userID)
	return handler(ctx, req)
}
