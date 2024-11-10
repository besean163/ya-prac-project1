package handlers

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

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
