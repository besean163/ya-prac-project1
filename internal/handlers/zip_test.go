package handlers

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"ya-prac-project1/internal/logger"

	"github.com/stretchr/testify/assert"
)

func TestNewZipWriter(t *testing.T) {
	w := httptest.NewRecorder()
	zipW := newZipWriter(w)

	assert.IsType(t, &zipWriter{}, zipW)
}

func TestWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	zipW := newZipWriter(w)

	zipW.WriteHeader(200)
}

func TestNewZipReader(t *testing.T) {
	r := io.NopCloser(strings.NewReader(""))
	zipW, _ := newZipReader(r)

	assert.IsType(t, &zipReader{}, zipW)
}

func TestRead(t *testing.T) {
	var buf bytes.Buffer

	w := gzip.NewWriter(&buf)
	w.Write([]byte("test"))
	w.Close()

	r := io.NopCloser(bytes.NewReader(buf.Bytes()))
	zipW, err := newZipReader(r)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 10)
	zipW.Read(b)
}

func TestClose(t *testing.T) {
	var buf bytes.Buffer

	w := gzip.NewWriter(&buf)
	w.Write([]byte("test"))
	w.Close()

	r := io.NopCloser(bytes.NewReader(buf.Bytes()))
	zipW, err := newZipReader(r)
	if err != nil {
		panic(err)
	}
	zipW.Close()
}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, world!"))
}

// Test gzip compression for response
func TestZipMiddleware_GzipCompression(t *testing.T) {
	// Create a test request with Accept-Encoding set to gzip
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with zipMiddleware
	handler := zipMiddleware(http.HandlerFunc(mockHandler))

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Get the response and check the Content-Encoding header
	resp := rr.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	// Check if the response body is gzip-compressed
	var buf bytes.Buffer
	_, err := io.Copy(&buf, resp.Body)
	assert.NoError(t, err)

	// Attempt to decompress the gzip response
	gzipReader, err := gzip.NewReader(&buf)
	assert.NoError(t, err)

	// Read and verify the decompressed body
	decompressedBody, err := io.ReadAll(gzipReader)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, world!", string(decompressedBody))
}

// Test gzip decompression for request body
func TestZipMiddleware_GzipDecompression(t *testing.T) {
	// Create a simple JSON body
	body := []byte(`{"name":"value"}`)

	// Compress the body using gzip
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(body)
	assert.NoError(t, err)
	gzipWriter.Close()

	// Create a test request with Content-Encoding set to gzip
	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with zipMiddleware
	handler := zipMiddleware(http.HandlerFunc(mockHandler))

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Check the response status code
	resp := rr.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test error during gzip decompression
func TestZipMiddleware_GzipDecompression_Error(t *testing.T) {
	logger.Set()
	// Create a malformed gzip request (not valid gzip data)
	body := []byte(`invalid gzip data`)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with zipMiddleware
	handler := zipMiddleware(http.HandlerFunc(mockHandler))

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Assert that the response is 500 (Internal Server Error)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// Test when no gzip compression is needed
func TestZipMiddleware_NoCompressionRequired(t *testing.T) {
	// Create a test request without gzip encoding
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "deflate") // Not gzip

	// Record the response
	rr := httptest.NewRecorder()

	// Wrap the mockHandler with zipMiddleware
	handler := zipMiddleware(http.HandlerFunc(mockHandler))

	// Call the handler with the request and recorder
	handler.ServeHTTP(rr, req)

	// Assert the response is normal without content encoding header
	resp := rr.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Content-Encoding"))
}
