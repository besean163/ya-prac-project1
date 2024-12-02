package handlers

import (
	"bytes"
	"compress/gzip"
	"errors"
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Content-Encoding"))
}

// Тест для zipWriter.Close
func TestZipWriterClose(t *testing.T) {
	// Используем httptest.ResponseRecorder, который реализует http.ResponseWriter
	rr := httptest.NewRecorder()

	// Создаём zipWriter, который будет писать в ResponseRecorder
	w := newZipWriter(rr)

	// Напишем данные в zipWriter
	data := []byte("Hello, World!")
	_, err := w.Write(data)
	if err != nil {
		t.Fatalf("unexpected error during Write: %v", err)
	}

	// Устанавливаем успешный статус код (200) для корректной работы заголовков
	w.WriteHeader(200)

	// Закрываем zipWriter
	err = w.Close()
	if err != nil {
		t.Fatalf("unexpected error during Close: %v", err)
	}

	// Проверяем, что в заголовках установлен правильный Content-Encoding
	if encoding := rr.Header().Get("Content-Encoding"); encoding != "gzip" {
		t.Fatalf("expected Content-Encoding to be gzip, got %s", encoding)
	}

	// Проверяем, что в теле ответа данные сжаты
	// Для этого создадим gzip.Reader для распаковки
	gzipReader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("unexpected error creating gzip reader: %v", err)
	}
	defer gzipReader.Close()

	// Читаем данные из gzip-архива
	decodedData, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("unexpected error reading from gzip reader: %v", err)
	}

	// Проверяем, что данные после распаковки совпадают с исходными
	if string(decodedData) != string(data) {
		t.Fatalf("expected %s, got %s", string(data), string(decodedData))
	}
}

type readCloserBuffer struct {
	*bytes.Buffer
}

func (r *readCloserBuffer) Close() error {
	// Для буфера Close не нужно ничего делать, можно просто вернуть nil
	return nil
}

type faultyReadCloser struct {
	*bytes.Buffer
	closeError error
}

func (f *faultyReadCloser) Close() error {
	return f.closeError
}

func TestZipReaderClose(t *testing.T) {
	// Создадим данные для сжатия
	data := []byte("Hello, World!")

	// Сжимаем данные в буфер
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	if err != nil {
		t.Fatalf("unexpected error during Write: %v", err)
	}
	zw.Close()

	// Оборачиваем буфер в readCloserBuffer, чтобы удовлетворить интерфейсу io.ReadCloser
	readCloser := &readCloserBuffer{Buffer: &buf}

	// Создаём новый zipReader
	zipR, err := newZipReader(readCloser)
	if err != nil {
		t.Fatalf("unexpected error creating zip reader: %v", err)
	}

	// Читаем данные из zipReader
	decodedData, err := io.ReadAll(zipR)
	if err != nil {
		t.Fatalf("unexpected error reading from zip reader: %v", err)
	}

	// Проверяем, что данные после распаковки совпадают с исходными
	if string(decodedData) != string(data) {
		t.Fatalf("expected %s, got %s", string(data), string(decodedData))
	}

	// Закрываем zipReader
	err = zipR.Close()
	if err != nil {
		t.Fatalf("unexpected error during Close: %v", err)
	}

	// Повторное закрытие zipReader не должно приводить к ошибке
	err = zipR.Close()
	if err != nil {
		t.Fatalf("unexpected error during second Close: %v", err)
	}
}

// Тест для zipReader.Close с ошибкой при закрытии базового ресурса
func TestZipReaderCloseWithFaultyCloser(t *testing.T) {
	// Создадим данные для сжатия
	data := []byte("Faulty Hello, World!")

	// Сжимаем данные в буфер
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	if err != nil {
		t.Fatalf("unexpected error during Write: %v", err)
	}
	zw.Close()

	// Оборачиваем буфер в faultyReadCloser, чтобы имитировать ошибку при закрытии
	closeError := errors.New("close error")
	faultyReadCloser := &faultyReadCloser{Buffer: &buf, closeError: closeError}

	// Создаём новый zipReader
	zipR, err := newZipReader(faultyReadCloser)
	if err != nil {
		t.Fatalf("unexpected error creating zip reader: %v", err)
	}

	// Закрываем zipReader
	err = zipR.Close()
	if err == nil || err.Error() != "close error" {
		t.Fatalf("expected close error, got %v", err)
	}
}

// Тест для zipReader, когда данные некорректны
func TestZipReaderWithInvalidData(t *testing.T) {
	// Создаём некорректные данные (не сжатые)
	data := []byte("This is not compressed data")

	// Оборачиваем данные в readCloserBuffer
	readCloser := &readCloserBuffer{Buffer: bytes.NewBuffer(data)}

	// Попробуем создать новый zipReader с некорректными данными
	_, err := newZipReader(readCloser)
	if err == nil {
		t.Fatalf("expected error when creating zip reader with invalid data, got nil")
	}
}
