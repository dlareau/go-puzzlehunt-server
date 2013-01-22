package main

import "net/http"

var mindex = Template("_base.html", "index.html")

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  check(mindex.Execute(w, nil))
}
