package main

import "compress/gzip"
import "net/http"
import "io"
import "strings"

type gzipResponseWriter struct {
  io.Writer
  http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
  return w.Writer.Write(b)
}

func GzipHandler(f http.Handler) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
      f.ServeHTTP(w, r)
      return
    }
    w.Header().Set("Content-Encoding", "gzip")
    gz := gzip.NewWriter(w)
    f.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
    gz.Close()
  })
}
