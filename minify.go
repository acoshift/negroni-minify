package minify

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/urfave/negroni"
)

// Minify is a middleware handler that minify html, css, js using tdewolff/minify
type Minify struct {
	*minify.M
}

// NewMinify returns new minify middleware
func NewMinify() *Minify {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	return &Minify{m}
}

func (m *Minify) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w := &responseWriter{
		w:    rw.(negroni.ResponseWriter),
		body: &bytes.Buffer{},
	}

	next(w, r)

	p := w.body.Bytes()
	if b, err := m.Bytes(rw.Header().Get("Content-Type"), p); err != nil {
		rw.WriteHeader(w.status)
		rw.Write(p)
	} else {
		rw.Header().Set("Content-Length", strconv.Itoa(len(b)))
		rw.WriteHeader(w.status)
		rw.Write(b)
	}
}

type responseWriter struct {
	w           negroni.ResponseWriter
	body        *bytes.Buffer
	status      int
	size        int
	beforeFuncs []func(negroni.ResponseWriter)
}

func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.callBefore()
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.body.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

func (rw *responseWriter) Before(before func(negroni.ResponseWriter)) {
	rw.beforeFuncs = append(rw.beforeFuncs, before)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	return rw.w.(http.CloseNotifier).CloseNotify()
}

func (rw *responseWriter) callBefore() {
	for i := len(rw.beforeFuncs) - 1; i >= 0; i-- {
		rw.beforeFuncs[i](rw)
	}
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
