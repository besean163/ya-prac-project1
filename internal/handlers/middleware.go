package handlers

import (
	"bytes"
	"crypto/hmac"
	random "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
				logger.Get().Info("reader create error", zap.String("error", err.Error()))
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
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		reqKey := r.Header.Get("HashSHA256")

		hash := hmac.New(sha256.New, []byte(key))
		hash.Write(bodyBytes)
		hashKey := hex.EncodeToString(hash.Sum(nil))

		if reqKey != "" && reqKey != hashKey {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hw := NewHashResponseWriter(w, key)
		h.ServeHTTP(hw, r)
	})
}

func cryptoKeyMiddleware(h http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key == "" {
			h.ServeHTTP(w, r)
			return
		}

		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			bodyBytes = decryptMessage(bodyBytes, key)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		h.ServeHTTP(w, r)
	})
}

func decryptMessage(buf []byte, cryptoKey string) []byte {
	if cryptoKey == "" {
		return buf
	}

	privKeyBytes, err := os.ReadFile(cryptoKey)
	if err != nil {
		fmt.Printf("can't read private key. Error: %s\n", err)
		return buf
	}

	block, _ := pem.Decode(privKeyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		fmt.Printf("failed to decode PEM block containing private key")
		return buf
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Printf("can't parse private key. Error: %s\n", err)
		return buf
	}

	decryptedMessage, err := rsa.DecryptPKCS1v15(random.Reader, privateKey, buf)
	if err != nil {
		fmt.Printf("can't encrypt message")
		return buf
	}

	return decryptedMessage
}

func allowedIPMiddleware(h http.Handler, trusted string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if trusted != "" {
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				http.Error(w, "Missing X-Real-IP header", http.StatusBadRequest)
				return
			}

			// Проверяем, разрешен ли этот IP
			if !isIPAllowed(clientIP, trusted) {
				http.Error(w, "Forbidden: IP not allowed", http.StatusForbidden)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

func isIPAllowed(ip, trusted string) bool {

	// Проверяем, входит ли IP в доверенную подсеть
	_, cidrNet, err := net.ParseCIDR(trusted)
	if err != nil {
		fmt.Println("Error parsing trusted subnet:", err)
		return false
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		fmt.Println("Error parsing client IP:", ip)
		return false
	}

	return cidrNet.Contains(clientIP)
}
