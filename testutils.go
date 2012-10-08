package main

import (
	"bytes"
	"net/http"
)

type MockResponseWriter struct {
	StatusCode int
	Buffer      *bytes.Buffer
	header      http.Header
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter { Buffer: new(bytes.Buffer), header: make(http.Header) }
}

func (w *MockResponseWriter) Header() http.Header {
	return w.header
}

func (w *MockResponseWriter) WriteHeader(c int) {
	w.StatusCode = c
}

func (w *MockResponseWriter) Write(b []byte) (int, error) {
	if w.StatusCode == 0 {
		w.StatusCode = http.StatusOK
	}
	return w.Buffer.Write(b)
}

