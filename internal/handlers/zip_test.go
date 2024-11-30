package handlers

import (
	"bytes"
	"compress/gzip"
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
