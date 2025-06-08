package internalhttp

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(log Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip := clientIP(r)
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.Info(fmt.Sprintf(
			"Request handled | IP=%s | Time=%s | Method=%s | Path=%s |"+
				" Proto=%s | Status=%s (%d) | Latency=%s | UserAgent=%s",
			ip,
			start.Format(time.RFC3339),
			r.Method,
			r.URL.Path,
			r.Proto,
			http.StatusText(rw.statusCode),
			rw.statusCode,
			duration.String(),
			userAgent(r),
		))
	})
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func userAgent(r *http.Request) string {
	ua := r.UserAgent()
	if ua == "" {
		return "-"
	}
	return ua
}
