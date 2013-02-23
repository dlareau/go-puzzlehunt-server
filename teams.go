package main

import "errors"
import "github.com/gorilla/mux"
import "labix.org/v2/mgo/bson"
import "net/http"

type Team struct {
  Id           bson.ObjectId "_id,omitempty"
  Name         string
  Phones       string
  Members      string
  EmailAddress string
  Err          error         ",omitempty"
  Username     string
  Password     string
}

var Teams = db.C("teams")

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
  check(AdminTemplate("teams/index.html").Execute(w, AllTeams()))
}

func TeamsCreate(w http.ResponseWriter, r *http.Request) {
  tcreate := AdminTemplate("teams/new.html", "teams/form.html")
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
  tedit := AdminTemplate("teams/edit.html", "teams/form.html")
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

func (t *Team) findName(name string) error {
  regex := bson.RegEx{ Pattern: "^" + name + "$", Options: "i" }
  return Teams.Find(bson.M{"username": regex}).One(t)
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
