// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"bufio"
	"io"
	"net"

	"github.com/nycu-ucr/gonet/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

var _ ResponseWriter = (*responseWriter)(nil)

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

func (w *responseWriter) WriteHeader(code int) {
	// println("gin/response_writer.go, responseWriter.WriteHeader Start")
	if code > 0 && w.status != code {
		if w.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
			return
		}
		w.status = code
	}
	// println("gin/response_writer.go, responseWriter.WriteHeader End")
}

func (w *responseWriter) WriteHeaderNow() {
	// println("gin/response_writer.go, responseWriter.WriteHeaderNow Start")
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
	// println("gin/response_writer.go, responseWriter.WriteHeaderNow End")
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	// println("gin/response_writer.go, responseWriter.Write Start")
	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	// println("gin/response_writer.go, responseWriter.Write End")
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	// println("gin/response_writer.go, responseWriter.WriteString Start")
	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	// println("gin/response_writer.go, responseWriter.WriteString End")
	return
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	// println("gin/response_writer.go, responseWriter.Flush Start")
	w.WriteHeaderNow()
	w.ResponseWriter.(http.Flusher).Flush()
	// println("gin/response_writer.go, responseWriter.Flush End")
}

func (w *responseWriter) Pusher() (pusher http.Pusher) {
	// println("gin/response_writer.go, responseWriter.Pusher Start")
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	// println("gin/response_writer.go, responseWriter.Pusher End")
	return nil
}
