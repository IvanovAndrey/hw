package grpc

import (
	"fmt"
	"net"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	logger     Logger
	app        Application
	cfg        *configuration.Config
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

func NewGrpcServer(cfg *configuration.Config, logger Logger, app Application) *Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(UnaryLoggingInterceptor(logger)),
	}

	if cfg.System.Grpc.ConnectionTimeout > 0 {
		opts = append(opts, grpc.ConnectionTimeout(time.Second*time.Duration(cfg.System.Grpc.ConnectionTimeout)))
	}

	srv := grpc.NewServer(opts...)
	proto.RegisterCalendarServer(srv, app)
	reflection.Register(srv)

	return &Server{
		grpcServer: srv,
		logger:     logger,
		app:        app,
		cfg:        cfg,
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf(":%d", s.cfg.System.Grpc.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.logger.Error("failed to listen: " + err.Error())
		return err
	}

	s.logger.Info("gRPC server listening on " + address)
	return s.grpcServer.Serve(listener)
}

func (s *Server) Stop() {
	s.logger.Info("stopping gRPC server...")
	s.grpcServer.GracefulStop()
}
