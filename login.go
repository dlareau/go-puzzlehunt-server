package main

import "encoding/hex"
import "github.com/alexcrichton/puzzlehunt/auth"
import "github.com/gorilla/securecookie"
import "net/http"
import "net/url"

var TokenSize = 24
var AdminToken string
var sc *securecookie.SecureCookie

func init() {
  AdminToken = hex.EncodeToString(securecookie.GenerateRandomKey(TokenSize))

  http.HandleFunc("/admin/auth", AdminLogin)
  http.HandleFunc("/teams/auth", TeamLogin)

  k1 := securecookie.GenerateRandomKey(TokenSize)
  k2 := securecookie.GenerateRandomKey(TokenSize)
  sc = securecookie.New(k1, k2)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
  _, given, err := auth.Basic(r)
  if err != nil || given != AdminPassword {
    auth.RequireAuth(w, r, AdminRealm)
    return
  }

  url := r.URL.Query().Get("back")
  cookie := &http.Cookie {
    Name: "admintoken",
    Value: AdminToken,
    Path: "/admin",
  }
  http.SetCookie(w, cookie)
  if url == "" {
    w.WriteHeader(http.StatusOK)
  } else {
    http.Redirect(w, r, url, http.StatusFound)
  }
}

func TeamLogin(w http.ResponseWriter, r *http.Request) {
  var team Team
  user, given, err := auth.Basic(r)
  if err == nil {
    err = team.findName(user)
  }
  if err != nil || team.Password != given {
    auth.RequireAuth(w, r, TeamRealm)
    return
  }

  encoded, err := sc.Encode("team", team.Id.Hex())
  check(err)

  url := r.URL.Query().Get("back")
  cookie := &http.Cookie {
    Name: "team",
    Value: encoded,
    Path: "/",
  }
  http.SetCookie(w, cookie)
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
      return
    }

    if r.Method == "GET" {
      path := "/admin/auth?back=" + url.QueryEscape(r.URL.String())
      http.Redirect(w, r, path, http.StatusFound)
    } else {
      w.WriteHeader(http.StatusUnauthorized)
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

    if r.Method == "GET" {
      path := "/teams/auth?back=" + url.QueryEscape(r.URL.String())
      http.Redirect(w, r, path, http.StatusFound)
    } else {
      w.WriteHeader(http.StatusUnauthorized)
    }
  })
}
