package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.status == 0 {
		lrw.status = http.StatusOK
	}
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.status = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func Logger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mLog := logger.With(zap.String("middleware", "logger"))
			start := time.Now()

			lrw := &loggingResponseWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(lrw, r)

			mLog.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status", lrw.status),
				zap.Int("size", lrw.size),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
