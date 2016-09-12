package gzip

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	*gzip.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}

	return w.Writer.Write(b)
}

// ExcludeRoute contains a route to exclude, supplied to New()
type ExcludeRoute string

// Middleware is created with New() and contains access to Handler function
type Middleware struct {
	excludeRoutes []ExcludeRoute
}

// New returns an instance of gzip.Middleware
func New(excludeRoutes []ExcludeRoute) *Middleware {
	return &Middleware{
		excludeRoutes: excludeRoutes,
	}
}

// Handler gzips responses
func (c *Middleware) Handler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		for _, e := range c.excludeRoutes {
			if strings.Contains(r.URL.Path, string(e)) {
				h.ServeHTTP(w, r)
				return
			}
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		gw := gzip.NewWriter(w)
		defer gw.Close()

		w = &gzipResponseWriter{gw, w}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
