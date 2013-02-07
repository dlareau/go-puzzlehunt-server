package main

import "github.com/gorilla/mux"
import "github.com/gorilla/schema"
import "labix.org/v2/mgo"
import "log"
import "net/http"
import _ "net/http/pprof"
import "os"
import "puzzlehunt/auth"
import "strings"

var mongo, db = opendb()
var decoder = schema.NewDecoder()

// type Handler func(http.ResponseWriter, *http.Request)
// type Handler = http.Handler

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

func H404(h http.Handler) http.Handler {
  return H(func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      err := recover()
      if err == mgo.ErrNotFound {
        http.Error(w, "not found", http.StatusNotFound)
      } else {
        panic(err)
      }
    }()
    h.ServeHTTP(w, r)
  })
}

func A404(h http.HandlerFunc) http.Handler {
  return H404(A(h))
}

func H(h http.HandlerFunc) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      e := recover()
      if e != nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorTemplate.Execute(w, e)
      }
    }()
    h.ServeHTTP(w, r)
  })
}

func A(h http.HandlerFunc) http.Handler {
  return Authenticate(h, AdminPassword, AdminRealm)
}

func TA(h func(http.ResponseWriter, *http.Request, *Team)) http.Handler {
  return H(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
  }))
}

func Authenticate(h http.Handler, password, realm string) http.Handler {
  return H(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    _, given, err := auth.Basic(r)
    if err != nil || given != password {
      auth.RequireAuth(w, r, realm)
    } else {
      h.ServeHTTP(w, r)
    }
  }))
}

func Log(handler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if !strings.HasPrefix(r.URL.String(), "/assets") {
      log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
    }
    handler.ServeHTTP(w, r)
  })
}

func main() {
  check(Puzzles.EnsureIndex(mgo.Index{ Key: []string{"slug"} }));
  check(Teams.EnsureIndex(mgo.Index{ Key: []string{"username"} }));
  check(Solutions.EnsureIndex(mgo.Index{ Key: []string{"teamid"} }));
  go Queue.Serve()
  go Progress.Serve()
  go PuzzleStatus.Serve()

  r := mux.NewRouter()

  r.Handle("/", H(HomeHandler)).Methods("GET")
  r.Handle("/map", TA(MapHandler)).Methods("GET")
  r.Handle("/map/puzzles/{id}", TA(MapPuzzleHandler)).Methods("GET", "POST")
  r.Handle("/map/{tag}/ws", PuzzleStatus.Endpoint())

  r.Handle("/admin/teams", A(TeamsIndex)).Methods("GET")
  r.Handle("/admin/teams/new", A(TeamsCreate)).Methods("GET", "POST")
  r.Handle("/admin/teams/{id}/edit", A404(TeamsEdit)).Methods("GET", "POST")
  r.Handle("/admin/teams/{id}/destroy", A404(TeamsDestroy)).Methods("POST")

  r.Handle("/admin/puzzles", A(PuzzlesIndex)).Methods("GET")
  r.Handle("/admin/puzzles/{puzzle_id}/unlock/{team_id}",
           H404(http.HandlerFunc(PuzzlesUnlock))).Methods("POST")
  r.Handle("/admin/puzzles/new", A(PuzzlesCreate)).Methods("GET", "POST")
  r.Handle("/admin/puzzles/{id}/edit", A404(PuzzlesEdit)).Methods("GET", "POST")
  r.Handle("/admin/puzzles/{id}/destroy", A404(PuzzlesDestroy)).Methods("POST")

  r.Handle("/admin/reset", A(ProgressReset)).Methods("POST")
  r.Handle("/admin/release", A(ProgressRelease)).Methods("POST")
  r.Handle("/admin", A(SubmissionsIndex)).Methods("GET")
  r.Handle("/admin/queue", A(SubmissionsIndex)).Methods("GET")
  r.Handle("/admin/queue/ws", Queue.Endpoint())
  r.Handle("/admin/progress", A(ProgressIndex)).Methods("GET")
  r.Handle("/admin/progress/ws", Progress.Endpoint())
  r.Handle("/admin/respond/{id}", A(SubmissionRespond)).Methods("POST")

  srv := http.FileServer(http.Dir("./assets"))
  http.Handle("/assets/", http.StripPrefix("/assets/", srv))
  http.Handle("/", r)

  log.Print("Serving requests...")
  check(http.ListenAndServe(ListenAddress, Log(http.DefaultServeMux)))
}
