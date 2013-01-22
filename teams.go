package main

import "errors"
import "github.com/gorilla/mux"
import "labix.org/v2/mgo/bson"
import "net/http"
import "net/mail"

type Team struct {
  Id      bson.ObjectId "_id,omitempty"
  Name    string
  Phones  string
  Members string
  EmailAddress string
  Err     error         ",omitempty"
}

var Teams   = db.C("teams")
var tindex  = AdminTemplate("teams/index.html")
var tcreate = AdminTemplate("teams/new.html", "teams/form.html")
var tedit   = AdminTemplate("teams/edit.html", "teams/form.html")

func AllTeams() []Team {
  teams := make([]Team, 0)
  var team Team
  iter := Teams.Find(nil).Sort("name").Iter()
  for iter.Next(&team) {
    teams = append(teams, team)
  }
  return teams
}

func TeamsIndex(w http.ResponseWriter, r *http.Request) {
  check(tindex.Execute(w, AllTeams()))
}

func TeamsCreate(w http.ResponseWriter, r *http.Request) {
  var team Team
  if r.Method == "GET" {
    check(tcreate.Execute(w, &team))
    return
  }
  team.Err = team.inherit(r)
  if team.Err == nil {
    team.Err = Teams.Insert(&team)
  }

  if team.Err != nil {
    check(tcreate.Execute(w, &team))
  } else {
    http.Redirect(w, r, "/admin/teams", http.StatusFound)
  }
}

func TeamsEdit(w http.ResponseWriter, r *http.Request) {
  var team Team
  team.find(mux.Vars(r)["id"])
  if r.Method == "GET" {
    check(tedit.Execute(w, &team))
    return
  }
  team.Err = team.inherit(r)
  if team.Err == nil {
    team.Err = Teams.UpdateId(team.Id, &team)
  }

  if team.Err != nil {
    check(tedit.Execute(w, &team))
  } else {
    http.Redirect(w, r, "/admin/teams", http.StatusFound)
  }
}

func TeamsDestroy(w http.ResponseWriter, r *http.Request) {
  check(Teams.RemoveId(bson.ObjectIdHex(mux.Vars(r)["id"])))
  http.Redirect(w, r, "/admin/teams", http.StatusFound)
}

func (t *Team) find(id string) {
  t.findId(bson.ObjectIdHex(id))
}

func (t *Team) findId(id bson.ObjectId) {
  check(Teams.FindId(id).One(t))
}

func (t *Team) inherit(r *http.Request) error {
  /* Parse from the form */
  err := r.ParseForm()
  if err != nil { return err }
  err = decoder.Decode(t, r.Form)
  if err != nil { return err }

  if t.Name == "" {
    return errors.New("Requires a name to be provided")
  }
  return nil
}

func (t *Team) Email() mail.Address {
  return mail.Address{Address: t.EmailAddress, Name: t.Name}
}

func (t *Team) UnlockMore() (bool, error) {
  /* Find the metapuzzle */
  var meta Puzzle
  err := Puzzles.Find(bson.M{"metapuzzle": true}).One(&meta)
  check(err)

  /* Find all currently unlocked puzzles by this team */
  var soln Solution
  iter := Solutions.Find(bson.M{"teamid": t.Id}).Iter()
  unlocked := make([]bson.ObjectId, 0)
  hasmeta := false
  solved := 0
  for iter.Next(&soln) {
    unlocked = append(unlocked, soln.PuzzleId)
    if soln.PuzzleId == meta.Id {
      hasmeta = true
    }
    if soln.SolvedAt.Year() > 2000 {
      solved++
    }
  }

  /* Find all puzzles that need to be unlocked, and unlock two more */
  iter2 := Puzzles.Find(bson.M{"metapuzzle":false, "secondround":false,
                               "_id": bson.M{"$nin": unlocked}}).
                   Sort("unlockidx").Limit(2).Iter()
  var puzzle Puzzle
  mailed := 0
  for iter2.Next(&puzzle) {
    _, err := CreateSolution(t, &puzzle)
    if err != nil {
      return false, err
    }
    mailed++
  }

  /* If 4 or more puzzles have been solved, then unlock the meta */
  if !hasmeta && solved >= 4 {
    _, err := CreateSolution(t, &meta)
    check(err)
    mailed++
  }

  return mailed > 0, nil
}
