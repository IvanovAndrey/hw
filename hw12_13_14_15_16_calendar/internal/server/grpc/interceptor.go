package grpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

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
