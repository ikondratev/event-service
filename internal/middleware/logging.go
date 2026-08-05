package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	size, err := w.ResponseWriter.Write(data)
	w.size += size

	return size, err
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			startedAt := time.Now()

			writer := &responseWriter{
				ResponseWriter: w,
			}

			logger.Debug(
				"request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			next.ServeHTTP(writer, r)

			duration := time.Since(startedAt)

			log := logger.With(
				"method", r.Method,
				"path", r.URL.Path,
				"status", writer.status,
				"response_size", writer.size,
				"duration_ms", duration.Milliseconds(),
			)

			switch {
			case writer.status >= http.StatusInternalServerError:
				log.Error("request completed")

			case writer.status >= http.StatusBadRequest:
				log.Warn("request completed")

			default:
				log.Info("request completed")
			}
		})
	}
}