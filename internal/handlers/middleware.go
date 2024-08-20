package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"
	"ya-prac-project1/internal/logger"

	"go.uber.org/zap"
)

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Get().Info(
			"get request",
			zap.String("method", r.Method),
			zap.String("uri", r.URL.Path),
		)

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
			"send response",
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

func hashKeyMiddleware(h http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqKey := r.Header.Get("HashSHA256")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Get().Debug("body read error", zap.String("error", err.Error()))
		}

		hash := hmac.New(sha256.New, []byte(key))
		hash.Write(body)
		hashKey := hex.EncodeToString(hash.Sum(nil))

		if reqKey != hashKey {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hw := NewHashResponseWriter(w, key)
		h.ServeHTTP(hw, r)
	})
}
