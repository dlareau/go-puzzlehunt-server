package main

import "encoding/hex"
import "github.com/alexcrichton/puzzlehunt/auth"
import "github.com/gorilla/securecookie"
import "net/http"
import "net/url"

var TokenSize = 24
var AdminToken string
var sc *securecookie.SecureCookie

/* re-key tokens/cookies across restarts so we don't have to persist anything */
func init() {
  AdminToken = hex.EncodeToString(securecookie.GenerateRandomKey(TokenSize))

  k1 := securecookie.GenerateRandomKey(TokenSize)
  k2 := securecookie.GenerateRandomKey(TokenSize)
  sc = securecookie.New(k1, k2)

  http.HandleFunc("/admin/auth", adminLogin)
  http.HandleFunc("/teams/auth", teamLogin)
}

func adminLogin(w http.ResponseWriter, r *http.Request) {
  _, given, err := auth.Basic(r)
  if err != nil || given != AdminPassword {
    auth.RequireAuth(w, r, AdminRealm)
  } else {
    setCookie(w, r, "admintoken", AdminToken, "/admin")
  }
}

func teamLogin(w http.ResponseWriter, r *http.Request) {
  var team Team
  user, given, err := auth.Basic(r)
  if err == nil {
    err = team.findName(user)
  }
  if err != nil || team.Password != given {
    auth.RequireAuth(w, r, TeamRealm)
  } else {
    encoded, err := sc.Encode("team", team.Id.Hex())
    check(err)
    setCookie(w, r, "team", encoded, "/")
  }
}

func setCookie(w http.ResponseWriter, r *http.Request, name, value, path string) {
  cookie := &http.Cookie {
    Name: name,
    Value: value,
    Path: path,
  }
  http.SetCookie(w, cookie)
  url := r.URL.Query().Get("back")
  if url == "" {
    w.WriteHeader(http.StatusOK)
  } else {
    http.Redirect(w, r, url, http.StatusFound)
  }
}

func AdminAuthenticate(h http.Handler) http.Handler {
  return H(func(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("admintoken")
    if err == nil && cookie.Value == AdminToken {
      h.ServeHTTP(w, r)
    } else {
      needAuth(w, r, "/admin/auth")
    }
  })
}

func TeamAuthenticate(h TeamHandler) http.Handler {
  return H(func(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("team")
    if err == nil {
      var id string
      err = sc.Decode("team", cookie.Value, &id)
      if err == nil {
        var t Team
        t.find(id)
        h(w, r, &t)
        return
      }
    }
    needAuth(w, r, "/teams/auth")
  })
}

func needAuth(w http.ResponseWriter, r *http.Request, callback string) {
  /* Only redirect GET requests because things like POST might have data
     associated with them which would be lost otherwise */
  if r.Method == "GET" {
    path := callback + "?back=" + url.QueryEscape(r.URL.String())
    http.Redirect(w, r, path, http.StatusFound)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
  }
}
