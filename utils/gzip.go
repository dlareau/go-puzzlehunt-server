package utils

import "compress/gzip"
import "io"
import "net/http"
import "strings"

type gzipResponseWriter struct {
  sniffed bool
  io.Writer
  http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
  if !w.sniffed {
    if w.ResponseWriter.Header().Get("Content-Type") == "" {
      w.ResponseWriter.Header().Set("Content-Type", http.DetectContentType(b))
    }
    w.sniffed = true
  }
  return w.Writer.Write(b)
}

func GzipHandler(f http.Handler) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") ||
       strings.Contains(r.Header.Get("Connection"), "Upgrade") {
      f.ServeHTTP(w, r)
      return
    }
    w.Header().Set("Content-Encoding", "gzip")
    gz := gzip.NewWriter(w)
    f.ServeHTTP(&gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
    gz.Close()
  })
}
