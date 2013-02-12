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
  /* If we don't sniff the uncompressed data, the http server will by default
     sniff our gzipped data, returning the wrong Content-Type. Hence, sniff the
     uncompressed data here the first time so we can set the right Content-Type
     based on the unzipped contents. */
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
    /* Don't gzip websocket connections because the current implementation of
       the http server/websocket server casts the ResponseWriter to a
       http.Hijacker which isn't implemented with the gzipResponseWriter */
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
