package internalhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/pkg/errors"
)

type Server struct {
	httpServer *http.Server
	logger     logger.Logger
	app        Application
}

type Application interface{}

func NewServer(cfg *configuration.Config, logger logger.Logger, app Application) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	httpServer := &http.Server{
		Addr:         cfg.System.HTTP.Address,
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  time.Duration(cfg.System.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.System.HTTP.WriteTimeout) * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		app:        app,
	}
}

func (s *Server) Start() error {
	go func() {
		s.logger.Info("Starting HTTP server at " + s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("HTTP server error: " + err.Error())
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctxWithTimeout)
}
