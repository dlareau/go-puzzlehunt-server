package main

import "github.com/gorilla/mux"
import "errors"
import "labix.org/v2/mgo/bson"
import "net/http"

type Puzzle struct {
  Id          bson.ObjectId "_id,omitempty"
  Name        string
  Url         string
  Answer      string
  Slug        string

  UnlockIdx   int     // when is this puzzle unlocked?
  Metapuzzle  bool    // is this the metapuzzle?

  Err         error         ",omitempty"
}

var Puzzles  = db.C("puzzles")
var pindex  = AdminTemplate("puzzles/index.html")
var pcreate = AdminTemplate("puzzles/new.html", "puzzles/form.html")
var pedit   = AdminTemplate("puzzles/edit.html", "puzzles/form.html")

func AllPuzzles() []Puzzle {
  puzzles := make([]Puzzle, 0)
  var puzzle Puzzle
  iter := Puzzles.Find(nil).Sort("metapuzzle", "unlockidx", "name").Iter()
  for iter.Next(&puzzle) {
    puzzles = append(puzzles, puzzle)
  }
  return puzzles
}

func PuzzlesIndex(w http.ResponseWriter, r *http.Request) {
  check(pindex.Execute(w, AllPuzzles()))
}

func PuzzlesCreate(w http.ResponseWriter, r *http.Request) {
  var puzzle Puzzle
  if r.Method == "GET" {
    check(pcreate.Execute(w, &puzzle))
    return
  }
  puzzle.Err = puzzle.inherit(r)
  if puzzle.Err == nil {
    puzzle.Err = Puzzles.Insert(&puzzle)
  }

  if puzzle.Err != nil {
    check(pcreate.Execute(w, &puzzle))
  } else {
    http.Redirect(w, r, "/admin/puzzles", http.StatusFound)
  }
}

func PuzzlesEdit(w http.ResponseWriter, r *http.Request) {
  var puzzle Puzzle
  puzzle.find(mux.Vars(r)["id"])
  if r.Method == "GET" {
    check(pedit.Execute(w, &puzzle))
    return
  }
  puzzle.Err = puzzle.inherit(r)
  if puzzle.Err == nil {
    puzzle.Err = Puzzles.UpdateId(puzzle.Id, &puzzle)
  }

  if puzzle.Err != nil {
    check(pedit.Execute(w, &puzzle))
  } else {
    http.Redirect(w, r, "/admin/puzzles", http.StatusFound)
  }
}

func PuzzlesDestroy(w http.ResponseWriter, r *http.Request) {
  check(Puzzles.RemoveId(bson.ObjectIdHex(mux.Vars(r)["id"])))
  http.Redirect(w, r, "/admin/puzzles", http.StatusFound)
}

func PuzzlesUnlock(w http.ResponseWriter, r *http.Request) {
  var puzzle Puzzle
  puzzle.find(mux.Vars(r)["puzzle_id"])
  var team Team
  team.find(mux.Vars(r)["team_id"])

  solution := Solution{TeamId: team.Id, PuzzleId: puzzle.Id}
  check(solution.Insert())
}

func (p *Puzzle) find(id string) {
  p.findId(bson.ObjectIdHex(id))
}

func (p *Puzzle) findId(id bson.ObjectId) {
  check(Puzzles.FindId(id).One(p))
}

func (p *Puzzle) findSlug(slug string) {
  check(Puzzles.Find(bson.M{"slug": slug}).One(p))
}

func (p *Puzzle) inherit(r *http.Request) error {
  /* Parse from the form */
  err := r.ParseForm()
  if err != nil { return err }
  err = decoder.Decode(p, r.Form)
  if err != nil { return err }
  if p.Name == "" {
    return errors.New("Requires a name to be provided")
  }
  p.Slug = Parameterize(p.Name);
  return nil
}

func (p *Puzzle) Classes() string {
  if p.Metapuzzle {
    return "meta"
  }
  return ""
}
