package main

import "bytes"
import "github.com/gorilla/mux"
import "errors"
import "labix.org/v2/mgo/bson"
import "text/template"
import "net/http"
import "net/mail"

import "puzzlehunt/email"

type Puzzle struct {
  Id          bson.ObjectId "_id,omitempty"
  Name        string
  Slug        string
  Url         string
  Answer      string

  SecondRound bool    // second round puzzles are all unlocked at once
  UnlockIdx   int     // first round puzzles unlocked in series
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
  iter := Puzzles.Find(nil).Sort("secondround", "metapuzzle",
                                 "unlockidx", "name").Iter()
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

  _, err := CreateSolution(&team, &puzzle)
  check(err)

  http.Redirect(w, r, "/admin/progress", http.StatusFound)
}

func (p *Puzzle) find(id string) {
  p.findId(bson.ObjectIdHex(id))
}

func (p *Puzzle) findId(id bson.ObjectId) {
  check(Puzzles.FindId(id).One(p))
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
  p.Slug = Parameterize(p.Name)

  /* Make sure the slug doesn't already exist */
  query := bson.M{"slug": p.Slug}
  if p.Id != "" {
    query["_id"] = bson.M{"$ne": p.Id}
  }
  cnt, err := Puzzles.Find(query).Count()
  if err != nil { return err }
  if cnt > 0 {
    return errors.New("Slug is already taken")
  }
  return nil
}

func (p *Puzzle) FromAddress() mail.Address {
  if p.SecondRound {
    return mail.Address{Address: Round2Username + "+" + p.Slug + "@" + EmailHost,
                        Name: Round2Name}
  }
  return mail.Address{Address: Round1Username + "+" + p.Slug + "@" + EmailHost,
                      Name: Round1Name}
}

var roundinit = template.Must(template.New("rinit").Parse(EmailInitialRound))
var round1 = template.Must(template.New("r1").Parse(EmailFirstRound))
var round2 = template.Must(template.New("r2").Parse(EmailSecondRound))
var roundmeta = template.Must(template.New("meta").Parse(EmailMetapuzzle))

func (p *Puzzle) EmailText(t *Team) string {
  desc := struct { Team *Team; Puzzle *Puzzle; From string }{t, p, Round1Name}
  buf := bytes.NewBufferString("")
  if p.Metapuzzle {
    check(roundmeta.Execute(buf, &desc))
  } else if p.SecondRound {
    desc.From = Round2Name
    check(round2.Execute(buf, &desc))
  } else if p.UnlockIdx <= 0 {
    check(roundinit.Execute(buf, &desc))
  } else {
    check(round1.Execute(buf, &desc))
  }
  return buf.String()
}

func (p *Puzzle) EmailTo(t *Team) error {
  var msg email.Message
  msg.From = p.FromAddress()
  msg.To = []mail.Address{t.Email()}
  msg.Subject = p.EmailSubject()
  msg.Content = p.EmailText(t)

  msg.Headers = map[string]string{"Content-Type": "text/plain; charset=utf8"}
  return msg.Send(MailServer)
}

func (p *Puzzle) UnlockedWithMeta() bool {
  return p.UnlockIdx >= 7
}

func (p *Puzzle) Classes() string {
  if p.SecondRound {
    return "second"
  } else if p.Metapuzzle {
    return "meta"
  }
  return ""
}
