package main

import "fmt"
import "labix.org/v2/mgo/bson"
import "net/http"
import "net/url"
import "strings"

var mindex = Template("_base.html", "index.html")
var mpasswords = Template("_base.html", "passwords/form.html")
var mreset = Template("_base.html", "passwords/reset.html")
var mfinal = Template("_base.html", "passwords/final.html")
var mdone = Template("_base.html", "passwords/done.html")

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  check(mindex.Execute(w, nil))
}

func PasswordReset(w http.ResponseWriter, r *http.Request) {
  check(r.ParseForm())
  data := struct { Msg string; Form url.Values; }{ "", r.Form }
  if r.Method == "GET" {
    check(mpasswords.Execute(w, data))
    return
  }

  for k, v := range ResetAnswers {
    arr := r.Form[k]
    if arr == nil || len(arr) != 1 || !strings.EqualFold(arr[0], v) {
      data.Msg = "One or more questions answered incorrectly!"
    }
  }
  if data.Msg == "" {
    check(mreset.Execute(w, r.Form))
  } else {
    check(mpasswords.Execute(w, data))
  }
}

func FinalQuestionnaire(w http.ResponseWriter, r *http.Request) {
  check(r.ParseForm())
  data := struct { Msg string; Form url.Values; }{ "", r.Form }
  if r.Method == "GET" {
    check(mfinal.Execute(w, data))
    return
  }

  var puzzle Puzzle
  correct := 0
  idx := 1
  iter := Puzzles.Find(bson.M{"metapuzzle":false, "secondround":true}).
                  Sort("unlockidx").Iter()
  for iter.Next(&puzzle) {
    submission := r.Form[fmt.Sprintf("a%d", idx)]
    if strings.EqualFold(submission[0], puzzle.Answer) {
      correct++
    }
    idx++
  }

  if correct >= 4 {
    check(mdone.Execute(w, nil))
  } else {
    data.Msg = "Need at least 4 correct answers"
    check(mfinal.Execute(w, data))
  }
}
