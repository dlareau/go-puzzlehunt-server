package main

import "github.com/gorilla/mux"
import "github.com/gorilla/schema"
import "labix.org/v2/mgo"
import "log"
import "net/http"
import _ "net/http/pprof"
import "os"
import "puzzlehunt/auth"

var mongo, db = opendb()
var decoder = schema.NewDecoder()

type Handler func(http.ResponseWriter, *http.Request)

const dbg = false

func opendb() (*mgo.Session, *mgo.Database) {
  if dbg {
    mgo.SetDebug(true)
    mgo.SetLogger(log.New(os.Stdout, "[mgo] ", log.LstdFlags))
  }
  session, err := mgo.Dial(MongoHost)
  check(err)
  return session, session.DB(MongoDatabase)
}

func check(err error) {
  if err != nil {
    panic(err)
  }
}

var errorTemplate = Template("_base.html", "error.html")

func H404(h Handler) Handler {
  return H(func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      err := recover()
      if err == mgo.ErrNotFound {
        http.Error(w, "not found", http.StatusNotFound)
      } else {
        panic(err)
      }
    }()
    h(w, r)
  })
}

func A404(h Handler) Handler {
  return H404(A(h))
}

func H(h Handler) Handler {
  return func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      e := recover()
      if e != nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorTemplate.Execute(w, e)
      }
    }()
    h(w, r)
  }
}

func A(h Handler) Handler {
  return Authenticate(h, AdminPassword, AdminRealm)
}

func TA(h func(http.ResponseWriter, *http.Request, *Team)) Handler {
  return H(func(w http.ResponseWriter, r *http.Request) {
    var team Team
    user, given, err := auth.Basic(r)
    if err == nil {
      err = team.findName(user)
    }
    if err != nil || team.Password != given {
      auth.RequireAuth(w, r, TeamRealm)
    } else {
      h(w, r, &team)
    }
  })
}

func Authenticate(h Handler, password, realm string) Handler {
  return H(func(w http.ResponseWriter, r *http.Request) {
    _, given, err := auth.Basic(r)
    if err != nil || given != password {
      auth.RequireAuth(w, r, realm)
    } else {
      h(w, r)
    }
  })
}

func main() {
  check(Puzzles.EnsureIndex(mgo.Index{ Key: []string{"slug"} }));
  check(Teams.EnsureIndex(mgo.Index{ Key: []string{"username"} }));

  r := mux.NewRouter()

  r.HandleFunc("/", H(HomeHandler)).Methods("GET")
  r.HandleFunc("/map", TA(MapHandler)).Methods("GET")
  r.HandleFunc("/map/puzzles/{id}", TA(MapPuzzleHandler)).Methods("GET", "POST")

  r.HandleFunc("/admin/teams", A(TeamsIndex)).Methods("GET")
  r.HandleFunc("/admin/teams/new", A(TeamsCreate)).Methods("GET", "POST")
  r.HandleFunc("/admin/teams/{id}/edit", A404(TeamsEdit)).Methods("GET", "POST")
  r.HandleFunc("/admin/teams/{id}/destroy", A404(TeamsDestroy)).Methods("POST")

  r.HandleFunc("/admin/puzzles", A(PuzzlesIndex)).Methods("GET")
  r.HandleFunc("/admin/puzzles/{puzzle_id}/unlock/{team_id}",
               H404(PuzzlesUnlock)).Methods("POST")
  r.HandleFunc("/admin/puzzles/new", A(PuzzlesCreate)).Methods("GET", "POST")
  r.HandleFunc("/admin/puzzles/{id}/edit", A404(PuzzlesEdit)).Methods("GET", "POST")
  r.HandleFunc("/admin/puzzles/{id}/destroy", A404(PuzzlesDestroy)).Methods("POST")

  r.HandleFunc("/admin/progress", A(ProgressIndex)).Methods("GET")
  r.HandleFunc("/admin/reset", A(ProgressReset)).Methods("POST")
  r.HandleFunc("/admin/release", A(ProgressRelease)).Methods("POST")
  r.HandleFunc("/admin", A(SubmissionsIndex)).Methods("GET")
  r.HandleFunc("/admin/queue", (SubmissionsIndex)).Methods("GET")
  r.HandleFunc("/admin/respond/{id}", A(SubmissionRespond)).Methods("POST")

  srv := http.FileServer(http.Dir("./assets"))
  http.Handle("/assets/", http.StripPrefix("/assets/", srv))
  http.Handle("/", r)

  log.Print("Serving requests...")
  check(http.ListenAndServe(ListenAddress, nil))
}
