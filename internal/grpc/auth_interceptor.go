package grpc

import (
	"context"
	"errors"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/golang-jwt/jwt/v4"
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

	userID, err := extractUserIDFromToken(authHeader[0])
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, userIDKey, userID)
	return handler(ctx, req)
}

func extractUserIDFromToken(tokenString string) (string, error) {
	type Claims struct {
		UserID string `json:"user_id"`
		jwt.RegisteredClaims
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(model.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims.UserID, nil
	}
	return "", errors.New("failed to extract user ID")
}
