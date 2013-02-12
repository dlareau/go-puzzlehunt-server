package utils

import "net/http"
import "time"

func CacheControl(handler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    expires := time.Now().Add(time.Second)
    w.Header().Add("Cache-Control", "public, max-age=1")
    w.Header().Add("Expires", expires.Format(time.RFC1123))
    handler.ServeHTTP(w, r)
  })
}
