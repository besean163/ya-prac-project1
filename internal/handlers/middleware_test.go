package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestIsIPAllowed(t *testing.T) {
	ip := "127.0.0.1"
	trusted := "127.0.0.1/32"
	assert.True(t, isIPAllowed(ip, trusted))

	assert.False(t, isIPAllowed("fail_value", trusted))
	assert.False(t, isIPAllowed(ip, "fail_value"))
}

func TestAllowedIPMiddleware_1(t *testing.T) {
	h := http.NewServeMux()
	n := allowedIPMiddleware(h, "")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	n.ServeHTTP(rec, req)
}

func TestAllowedIPMiddleware_2(t *testing.T) {
	h := http.NewServeMux()
	n := allowedIPMiddleware(h, "127.0.0.1/32")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	n.ServeHTTP(rec, req)
}

func TestAllowedIPMiddleware_3(t *testing.T) {
	h := http.NewServeMux()
	n := allowedIPMiddleware(h, "127.0.0.1/32")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "127.0.0.1")
	rec := httptest.NewRecorder()
	n.ServeHTTP(rec, req)
}

// Test that the middleware correctly verifies the hash
func TestHashKeyMiddleware_ValidHash(t *testing.T) {
	// Body of the request
	body := []byte("test body")

	// The secret key used to generate the hash
	key := "secret"

	// Create the HMAC hash of the body using the key
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(body)
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	// Create the test request with the correct HashSHA256 header
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", expectedHash)

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with hashKeyMiddleware
	handler := hashKeyMiddleware(http.HandlerFunc(mockHandler), key)

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Assert the response status code is OK
	assert.Equal(t, http.StatusOK, rr.Code)
}

// Test that the middleware returns BadRequest for invalid hash
func TestHashKeyMiddleware_InvalidHash(t *testing.T) {
	// Body of the request
	body := []byte("test body")

	// The secret key used to generate the hash
	key := "secret"

	// Create a wrong hash (alter the correct hash)
	wrongHash := "wronghash"

	// Create the test request with an invalid HashSHA256 header
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", wrongHash)

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with hashKeyMiddleware
	handler := hashKeyMiddleware(http.HandlerFunc(mockHandler), key)

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Assert the response status code is BadRequest
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// Test that the middleware works when no HashSHA256 header is provided
func TestHashKeyMiddleware_NoHashHeader(t *testing.T) {
	// Body of the request
	body := []byte("test body")

	// The secret key used to generate the hash
	key := "secret"

	// Create the test request without HashSHA256 header
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with hashKeyMiddleware
	handler := hashKeyMiddleware(http.HandlerFunc(mockHandler), key)

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Assert the response status code is OK
	assert.Equal(t, http.StatusOK, rr.Code)
}

// Test that the middleware sets the HashSHA256 header in the response correctly
func TestHashKeyMiddleware_ResponseHash(t *testing.T) {
	// Body of the request
	body := []byte("test body")

	// The secret key used to generate the hash
	key := "secret"

	// Create the test request without HashSHA256 header (testing response hash)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with hashKeyMiddleware
	handler := hashKeyMiddleware(http.HandlerFunc(mockHandler), key)

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Expected response body and hash
	expectedBody := "Hello, world!"
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(expectedBody))
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	// Check if the response has the correct HashSHA256 header
	assert.Equal(t, expectedHash, rr.Header().Get("HashSHA256"))
}

func generateRSAKey(t *testing.T) string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Write private key to temp file
	privKeyFile, err := ioutil.TempFile("", "private_key_")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer privKeyFile.Close()

	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	err = pem.Encode(privKeyFile, block)
	if err != nil {
		t.Fatalf("failed to write private key to file: %v", err)
	}

	return privKeyFile.Name()
}

// Test decryptMessage when cryptoKey is empty
func TestDecryptMessage_NoKey(t *testing.T) {
	// Message to be decrypted (in this case we just return it as is)
	buf := []byte("some encrypted message")
	cryptoKey := ""

	// Call decryptMessage with no key
	result := decryptMessage(buf, cryptoKey)

	// Assert that the result is the same as input
	assert.Equal(t, buf, result)
}

// Test decryptMessage with a valid key
func TestDecryptMessage_ValidKey(t *testing.T) {
	// Generate RSA private key and file
	cryptoKey := generateRSAKey(t)
	defer os.Remove(cryptoKey) // Cleanup after test

	// Generate public key for encryption
	privKeyBytes, err := ioutil.ReadFile(cryptoKey)
	if err != nil {
		t.Fatalf("can't read private key: %v", err)
	}
	block, _ := pem.Decode(privKeyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Fatalf("failed to decode PEM block containing private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("can't parse private key: %v", err)
	}

	// Encrypt a message using the private key
	message := []byte("Hello, RSA encryption!")
	encryptedMessage, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, message)
	if err != nil {
		t.Fatalf("failed to encrypt message: %v", err)
	}

	// Call decryptMessage with the valid private key
	decryptedMessage := decryptMessage(encryptedMessage, cryptoKey)

	// Assert that the decrypted message matches the original
	assert.Equal(t, message, decryptedMessage)
}

// Test decryptMessage with an invalid key file
func TestDecryptMessage_InvalidKeyFile(t *testing.T) {
	// Pass a non-existent key file
	cryptoKey := "non_existent_key.pem"

	// Message to be decrypted (it should not be decrypted as the key is invalid)
	buf := []byte("some encrypted message")

	// Call decryptMessage with the invalid key
	result := decryptMessage(buf, cryptoKey)

	// Assert that the result is the same as input
	assert.Equal(t, buf, result)
}

// Test decryptMessage with an invalid key format
func TestDecryptMessage_InvalidKeyFormat(t *testing.T) {
	// Create a temp file with invalid key content
	invalidKeyFile, err := ioutil.TempFile("", "invalid_key_")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer invalidKeyFile.Close()

	// Write invalid content to the file
	invalidKeyFile.WriteString("invalid key content")

	// Call decryptMessage with the invalid key
	result := decryptMessage([]byte("some encrypted message"), invalidKeyFile.Name())

	// Assert that the result is the same as input
	assert.Equal(t, []byte("some encrypted message"), result)
}
