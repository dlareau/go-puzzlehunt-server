package main

import "net/http"

var mindex = Template("_base.html", "index.html")

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  check(mindex.Execute(w, nil))
}

func MapHandler(w http.ResponseWriter, r *http.Request, t *Team) {
  check(mindex.Execute(w, nil))
}
