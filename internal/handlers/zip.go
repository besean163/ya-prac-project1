package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
)

type zipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newZipWriter(w http.ResponseWriter) *zipWriter {
	return &zipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (zipW *zipWriter) Header() http.Header {
	return zipW.w.Header()
}

func (zipW *zipWriter) Write(p []byte) (int, error) {
	return zipW.zw.Write(p)
}

func (zipW *zipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		zipW.w.Header().Set("Content-Encoding", "gzip")
	}
	zipW.w.WriteHeader(statusCode)
}

func (zipW *zipWriter) Close() error {
	return zipW.zw.Close()
}

type zipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newZipReader(r io.ReadCloser) (*zipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &zipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (zipR zipReader) Read(p []byte) (n int, err error) {
	return zipR.zr.Read(p)
}

func (zipR *zipReader) Close() error {
	if err := zipR.r.Close(); err != nil {
		return err
	}

	return zipR.zr.Close()
}
