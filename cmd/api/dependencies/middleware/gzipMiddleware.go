package middleware

import (
    "compress/gzip"
    "net/http"
    "strings"
)

func (a *MiddlewareDependencies)GzipResponseMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Only compress if client accepts gzip
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }

        // Wrap the ResponseWriter
        gz := gzip.NewWriter(w)
        defer gz.Close()

        w.Header().Set("Content-Encoding", "gzip")
        w.Header().Set("Vary", "Accept-Encoding")

        gzrw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
        next.ServeHTTP(gzrw, r)
    })
}

type gzipResponseWriter struct {
    http.ResponseWriter
    Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}

func (a *MiddlewareDependencies) GzipRequestMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Content-Encoding") == "gzip" {
            gz, err := gzip.NewReader(r.Body)
            if err != nil {
                http.Error(w, "Invalid gzip body", http.StatusBadRequest)
                return
            }
            defer gz.Close()
            r.Body = gz
        }
        next.ServeHTTP(w, r)
    })
}