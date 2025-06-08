package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
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

func UnaryLoggingInterceptor(log Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		var ip string

		if p, ok := peer.FromContext(ctx); ok {
			ip, _, _ = net.SplitHostPort(p.Addr.String())
		}
		if ip == "" {
			ip = "unknown"
		}

		userAgent := getUserAgent(ctx)

		resp, err = handler(ctx, req)
		code := status.Code(err)
		latency := time.Since(start)

		log.Info(
			"gRPC request handled | " +
				"IP=" + ip + " | " +
				"Time=" + start.Format(time.RFC3339) + " | " +
				"Method=" + info.FullMethod + " | " +
				"Code=" + code.String() + " | " +
				"Latency=" + latency.String() + " | " +
				"UserAgent=" + userAgent,
		)

		return resp, err
	}
}

func getUserAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "-"
	}
	ua := md.Get("user-agent")
	if len(ua) == 0 {
		return "-"
	}
	return ua[0]
}
