package handlers

import (
	"net/http"
	"strings"
	"time"
	"ya-prac-project1/internal/logger"

	"go.uber.org/zap"
)

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ld := LogData{}
		lResponseWriter := &LogResponse{
			ResponseWriter: w,
			Data:           ld,
		}
		start := time.Now()
		h.ServeHTTP(lResponseWriter, r)
		duration := time.Since(start)

		lResponseWriter.Data.Method = r.Method
		lResponseWriter.Data.URI = r.RequestURI

		logger.Get().Info(
			"get request",
			zap.String("method", lResponseWriter.Data.Method),
			zap.String("uri", lResponseWriter.Data.URI),
			zap.Int("status", lResponseWriter.Data.Status),
			zap.Int("size", lResponseWriter.Data.Size),
			zap.Duration("time", duration),
		)
	})
}

func zipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		if supportGzip {
			w.Header().Set("Content-Encoding", "gzip")
			cw := newZipWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentType := r.Header.Get("Content-Type")
		contentEncoding := r.Header.Get("Content-Encoding")
		sendGzip := strings.Contains(contentEncoding, "gzip")
		if sendGzip && (contentType == "application/json" || contentType == "text/html") {
			cr, err := newZipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	})
}
