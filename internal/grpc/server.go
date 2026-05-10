package grpc

import (
	"fmt"
	"net"

	pb "github.com/amfib87/go-musthave-shortener-tpl/proto"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/handler"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedShortenerServiceServer
	Handler *handler.Handler
}

func NewServer(handler *handler.Handler) *Server {
	return &Server{Handler: handler}
}

func StartGRPCServer(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed net.Listen: %w", err)
	}

	s := grpc.NewServer()
	pb.RegisterShortenerServiceServer(s, &Server{})

	return s.Serve(lis)
}
