package minify

import (
	"bytes"
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
	w := &writer{
		ResponseWriter: rw.(negroni.ResponseWriter),
		Body:           &bytes.Buffer{},
	}

	next(negroni.NewResponseWriter(w), r)

	hd := rw.Header()
	p := w.Body.Bytes()
	b, err := m.Bytes(hd.Get("Content-Type"), p)
	if err != nil {
		rw.Write(p)
	} else {
		hd.Del("Content-Length")
		hd.Set("Content-Length", strconv.Itoa(len(b)))
		rw.Write(b)
	}
}

type writer struct {
	negroni.ResponseWriter
	Body *bytes.Buffer
}

func (w *writer) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}

	w.Body.Write(b)

	return len(b), nil
}
