package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"net/http"
)

type HashResponseWriter struct {
	http.ResponseWriter
	hash hash.Hash
}

func NewHashResponseWriter(w http.ResponseWriter, key string) HashResponseWriter {
	return HashResponseWriter{
		ResponseWriter: w,
		hash:           hmac.New(sha256.New, []byte(key)),
	}
}

func (hw HashResponseWriter) Write(b []byte) (int, error) {
	hw.hash.Write(b)
	hashKey := hex.EncodeToString(hw.hash.Sum(nil))
	hw.Header().Set("HashSHA256", hashKey)

	return hw.ResponseWriter.Write(b)
}
