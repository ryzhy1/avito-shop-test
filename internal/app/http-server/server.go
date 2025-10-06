package http_server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	log          *slog.Logger
	address      string
	handler      *gin.Engine
	httpServer   *http.Server
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewServer(log *slog.Logger, address string, handler *gin.Engine) *Server {
	return &Server{
		log:          log,
		address:      address,
		handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func (s *Server) MustRun() {
	if err := s.Run(); err != nil {
		panic(err)
	}
}

func (s *Server) Run() error {
	const op = "HTTPServer.Run"

	log := s.log.With(
		slog.String("op", op),
		slog.String("port", s.address),
	)

	log.Info("HTTP http-server started")

	if err := s.handler.Run(s.address); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	const op = "HTTPServer.Stop"

	s.log.With(slog.String("op", op)).
		Info("HTTP http-server stopped")

	return s.httpServer.Shutdown(ctx)
}
