package internalhttp

import (
	"net"
	"net/http"
	"strconv"
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

		log.Info(
			"Request handled | " +
				"IP=" + ip + " | " +
				"Time=" + start.Format(time.RFC3339) + " | " +
				"Method=" + r.Method + " | " +
				"Path=" + r.URL.Path + " | " +
				"Proto=" + r.Proto + " | " +
				"Status=" + http.StatusText(rw.statusCode) + " (" + strconv.Itoa(rw.statusCode) + ") | " +
				"Latency=" + duration.String() + " | " +
				"UserAgent=" + userAgent(r),
		)
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
