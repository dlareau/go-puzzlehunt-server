package auth

import "errors"
import "encoding/base64"
import "net/http"
import "strings"

func Basic(r *http.Request) (string, string, error) {
  /* Extract the "Authorization: ..." header */
  auth := r.Header.Get("Authorization")
  if auth == "" {
    return "", "", errors.New("No authorization header provided")
  }

  /* Must be of the form "Basic ..." */
  parts := strings.SplitN(auth, " ", 2)
  if len(parts) != 2 || parts[0] != "Basic" {
    return "", "", errors.New("Basic authentication not provided")
  }

  /* The extra should be a base64 encoding of 'user:pass' */
  bytes, err := base64.StdEncoding.DecodeString(parts[1])
  if err != nil {
    return "", "", err
  }
  parts = strings.SplitN(string(bytes), ":", 2)
  if len(parts) != 2 {
    return "", "", errors.New("Invalid authorization header provided")
  }

  /* We found it! */
  return parts[0], parts[1], nil
}

func RequireAuth(w http.ResponseWriter, r *http.Request, realm string) {
  w.Header().Set("WWW-Authenticate", `Basic realm="` + realm + `"`)
  w.WriteHeader(http.StatusUnauthorized)
}
