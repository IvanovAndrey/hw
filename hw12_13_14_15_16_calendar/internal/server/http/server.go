package internalhttp

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type Server struct {
	httpServer *http.Server
	logger     Logger
	app        Application
}

type Application interface {
	proto.CalendarServer
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

func NewHTTPServer(cfg *configuration.Config, logger Logger, app Application) *Server {
	gw := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}))

	if err := proto.RegisterCalendarHandlerFromEndpoint(
		context.Background(),
		gw,
		"127.0.0.1:"+strconv.Itoa(int(cfg.System.Grpc.Port)),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}); err != nil {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle("/", gw)

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
