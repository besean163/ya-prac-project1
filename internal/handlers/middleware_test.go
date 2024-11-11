package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecryptMessage(t *testing.T) {
	decryptMessage([]byte{}, "")
}
func TestDecryptMessage2(t *testing.T) {
	decryptMessage([]byte{}, "testdata/private.pem")
}

func TestDecryptMessage3(t *testing.T) {
	decryptMessage([]byte{}, "testdata/wrong_private.pem")
}

func TestHashKeyMiddleware(t *testing.T) {
	h := http.NewServeMux()
	hashKeyMiddleware(h, "")
}

func TestHashKeyMiddleware2(t *testing.T) {
	h := http.NewServeMux()
	n := hashKeyMiddleware(h, "secret")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	n.ServeHTTP(httptest.NewRecorder(), req)
}

func TestCryptoKeyMiddleware(t *testing.T) {
	h := http.NewServeMux()
	n := cryptoKeyMiddleware(h, "secret")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	n.ServeHTTP(httptest.NewRecorder(), req)
}

func TestZipMiddleware(t *testing.T) {
	h := http.NewServeMux()
	n := zipMiddleware(h)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	n.ServeHTTP(httptest.NewRecorder(), req)
}
