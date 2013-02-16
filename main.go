package main

import "github.com/alexcrichton/puzzlehunt/utils"
import "github.com/gorilla/mux"
import "github.com/gorilla/schema"
import "labix.org/v2/mgo"
import "log"
import "net"
import "net/http"
import _ "net/http/pprof"
import "os"
import "os/signal"
import "strings"
import "time"

var db = opendb()
var decoder = schema.NewDecoder()

const dbg = false

type TeamHandler func(w http.ResponseWriter, r *http.Request, t *Team)

func opendb() *mgo.Database {
  if dbg {
    mgo.SetDebug(true)
    mgo.SetLogger(log.New(os.Stdout, "[mgo] ", log.LstdFlags))
  }
  session, err := mgo.Dial(MongoHost)
  check(err)
  return session.DB(MongoDatabase)
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
        log.Printf("internal error %s", e)
        errorTemplate.Execute(w, e)
      }
    }()
    h(w, r)
  })
}

func A(h http.HandlerFunc) http.Handler {
  return AdminAuthenticate(h)
}

func Log(handler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    url := r.URL.String() /* may change depending on routing, so save before */
    handler.ServeHTTP(w, r)
    if !strings.HasPrefix(url, "/assets") && url != "/favicon.ico" &&
       !strings.HasSuffix(url, "/ws") {
      log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, url, time.Since(start))
    }
  })
}

func main() {
  /* If this is ot just precompile some assets, do that up front and return
     quickly */
  if len(os.Args) > 1 && os.Args[1] == "precompile" {
    PasteServer.Config().Compressed = true
    err := PasteServer.Compile("precompiled")
    if err != nil {
      panic(err)
    }
    return
  }

  check(Puzzles.EnsureIndex(mgo.Index{ Key: []string{"slug"} }));
  check(Teams.EnsureIndex(mgo.Index{ Key: []string{"username", "name"} }));
  check(Solutions.EnsureIndex(mgo.Index{ Key: []string{"teamid", "receivedat"} }));

  /* spawn off each of the threads responsible for managing websockets */
  go Queue.Serve()
  go Progress.Serve()
  go PuzzleStatus.Serve()

  /* Routing tables! */
  r := mux.NewRouter()
  TA := TeamAuthenticate

  r.Handle("/", TA(MapHandler)).Methods("GET")
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

  http.Handle("/assets/", http.StripPrefix("/assets", PasteServer))
  http.Handle("/favicon.ico", PasteServer)
  http.Handle("/", r)

  log.Print("Serving requests...")
  listen := ListenAddress
  if len(os.Args) > 1 {
    listen = os.Args[1]
  }

  l, err := net.Listen("tcp", listen)
  check(err)

  /* Be sure we can run code after the server exits */
  exit := make(chan os.Signal)
  signal.Notify(exit, os.Interrupt, os.Kill)
  go func() {
    <-exit
    l.Close()
  }()

  http.Serve(l, Log(utils.GzipHandler(http.DefaultServeMux)))

  println("Waiting for all problems to be marked as correct")
  CorrectNotifiers.Wait()
}
