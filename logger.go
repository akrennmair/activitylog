package main

import (
	"log"
	"net/http"
	"time"
)

type LoggingHandler struct {
	h http.Handler
}

type LogResponseWriter struct {
	http.ResponseWriter
	RespCode int
	Size     int
}

func (w *LogResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *LogResponseWriter) Write(data []byte) (s int, err error) {
	s, err = w.ResponseWriter.Write(data)
	w.Size += s
	return
}

func (w *LogResponseWriter) WriteHeader(r int) {
	w.ResponseWriter.WriteHeader(r)
	w.RespCode = r
}

func Logger(h http.Handler) http.Handler {
	return &LoggingHandler{h: h}
}

func (h *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lrw := &LogResponseWriter{ResponseWriter: w}
	t := time.Now()
	h.h.ServeHTTP(lrw, r)
	duration := time.Since(t).String()
	if lrw.RespCode == 0 {
		lrw.RespCode = 200
	}
	log.Printf("Request: %s \"%s %s %s\" %d %d (%s)", r.RemoteAddr, r.Method, r.RequestURI, r.Proto, lrw.RespCode, lrw.Size, duration)
}
