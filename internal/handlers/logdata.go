package handlers

import "net/http"

type LogData struct {
	URI    string
	Method string
	Status int
	Size   int
}

type LogResponse struct {
	http.ResponseWriter
	Data LogData
}

func (r *LogResponse) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Data.Size += size
	return size, err
}

func (r *LogResponse) WriteHeader(statusCode int) {
	// vet говорит что эта строка лишняя, не понимаю почему
	r.ResponseWriter.WriteHeader(statusCode)
	r.Data.Status = statusCode
}
