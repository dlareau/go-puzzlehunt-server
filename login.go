/**
 * Implementation of logging users in for both teams and the admin portion of
 * the site.
 *
 * This uses HTTP header-based authentication (the Authentication header) for
 * handling logins/passwords. This is insecure for man in the middle attacks
 * unless we're using SSL which we're not. This is a small enough
 * website/application that I don't really care, however.
 *
 * Most of the site relies on cookies to ensure that someone is authenticated.
 * If a cookie is not found for a request, the user is redirected to an
 * authorization page which requests the 'Authorization' header via the
 * 'WWW-Authorize' header (http basic authentication). Once received and
 * validated, then cookies are issued in two situations:
 *
 *      1. Admins all receive a common and constant 'AdminToken' under the
 *         cookie name 'admintoken' which is only sent for the '/admin' paths.
 *      2. Teams are issued an encrypted version of their team id so they can't
 *         easily spoof another team. This is stored in the 'team' cookie and
 *         sent on all requests.
 *
 * The constant admin token and encryption keys for the team's cookies are
 * generated at the start of the program, so between instances of a server users
 * have to re-login. Normally browsers can handle this transparently, but
 * there's a hiccup with anything which isn't a POST request. A GET can be
 * easily redirected around, but a POST contains data which will be lsot through
 * a series of redirects. To mitigate this, we've got JS running to catch
 * authorization errors on form submissions which then re-authenticate manually,
 * afterwards resubmitting the form.
 *
 * N.B. This was all done after noticing that Chrome refused to cache anything
 * with an Authorization header. It was later realized that this may be a bug in
 * Chrome, and I never went back and tested whether this was actually the case.
 * Oh well, water under the bridge or something like that?
 */

package main

import "encoding/hex"
import "github.com/alexcrichton/puzzlehunt/auth"
import "github.com/gorilla/securecookie"
import "net/http"
import "net/url"
import "strings"

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
  if err != nil || !strings.EqualFold(team.Password, given) {
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
