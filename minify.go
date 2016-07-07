package minify

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

var (
	m *minify.M
)

func init() {
	m = minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
}

type writer struct {
	http.ResponseWriter
	Body        *bytes.Buffer
	code        int
	wroteHeader bool
}

func (w *writer) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *writer) WriteHeader(code int) {
	if !w.wroteHeader {
		w.code = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *writer) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.Body != nil {
		w.Body.Write(b)
	}

	return len(b), nil
}

// Minify Middleware
func Minify(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := &writer{
			ResponseWriter: w,
			Body:           &bytes.Buffer{},
		}

		h.ServeHTTP(wr, r)

		hd := w.Header()
		p := wr.Body.Bytes()
		b, err := m.Bytes(hd.Get("Content-Type"), p)
		if err != nil {
			w.Write(p)
		} else {
			hd.Del("Content-Length")
			hd.Set("Content-Length", strconv.Itoa(len(b)))

			w.Write(b)
		}
	})
}
