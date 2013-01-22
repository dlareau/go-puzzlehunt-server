package main

import "github.com/gorilla/mux"
import "labix.org/v2/mgo/bson"
import "net/http"
import "strings"
import "time"

var mindex = Template("_base.html", "index.html")
var mpuzzle = Template("_base.html", "puzzle.html")
var mpuzzles = Template("_base.html", "puzzles.html")

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  check(mindex.Execute(w, nil))
}

func MapHandler(w http.ResponseWriter, r *http.Request, t *Team) {
  check(mpuzzles.Execute(w, AllPuzzles()))
}

func MapPuzzleHandler(w http.ResponseWriter, r *http.Request, t *Team) {
  var puzzle Puzzle
  puzzle.findSlug(mux.Vars(r)["id"]);
  var soln Solution
  check(Solutions.Find(bson.M{"teamid": t.Id, "puzzleid": puzzle.Id}).One(&soln))

  if r.Method != "GET" {
    check(r.ParseForm())
    answer := r.Form["answer"][0]
    submission := Submission { SolutionId: soln.Id,
                               TeamName: t.Name,
                               PuzzleName: puzzle.Name,
                               Answer: answer,
                               ReceivedAt: time.Now() }
    if answer == puzzle.Answer {
      submission.Status = CorrectUnreplied
    } else if strings.Index(answer, " ") != -1 {
      submission.Status = InvalidAnswer
    } else {
      submission.Status = IncorrectUnreplied
    }
    check(Submissions.Insert(&submission))
  }

  submissions := make([]Submission, 0)
  var s Submission
  iter := Submissions.Find(bson.M{"solutionid": soln.Id}).
                      Sort("-receivedat").Iter()
  for iter.Next(&s) {
    submissions = append(submissions, s)
  }

  data := struct {
    Puzzle *Puzzle
    Submissions []Submission
  }{&puzzle, submissions}

  check(mpuzzle.Execute(w, &data))
}

func AnswerStatus(s Submission) string {
  switch (s.Status) {
    case CorrectUnreplied, IncorrectUnreplied: return "validating...";
    case Correct:                              return "correct";
    case InvalidAnswer:                        return InvalidAnswerText;
    case IncorrectReplied:                     return "incorrect: " + s.Comment;
  }
  return ""
}
