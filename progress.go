package main

import "github.com/gorilla/mux"
import "labix.org/v2/mgo/bson"
import "net/http"
import "time"

/* A solution can be solved or possibly not. This is a placeholder for whether a
   team has unlocked a puzzle, and then for whether the team has solved the
   puzzle or not (nonzero SolvedAt) */
type Solution struct {
  Id        bson.ObjectId "_id,omitempty"
  TeamId    bson.ObjectId
  PuzzleId  bson.ObjectId
  SolvedAt  time.Time
}

type Submission struct {
  Id          bson.ObjectId "_id,omitempty"
  SolutionId  bson.ObjectId
  TeamName    string
  PuzzleName  string
  Answer      string
  Status      SubmissionStatus
  Comment     string
  ReceivedAt  time.Time
}

type SubmissionStatus int

const (
  Correct SubmissionStatus = iota
  CorrectUnreplied
  InvalidAnswer
  IncorrectReplied
  IncorrectUnreplied
)

type SolutionList []Solution
type SolutionMap  map[bson.ObjectId]Solution
type TeamMap      map[bson.ObjectId]Team
type PuzzleMap    map[bson.ObjectId]Puzzle

var Solutions = db.C("solutions")
var Submissions = db.C("submissions")

var solutionst = AdminTemplate("progress/solutions.html")
var queuet = AdminTemplate("progress/queue.html")

func AllSolutions() []Solution {
  solutions := make([]Solution, 0)
  var solution Solution
  for iter := Solutions.Find(nil).Iter(); iter.Next(&solution); {
    solutions = append(solutions, solution)
  }
  return solutions
}

func AllSubmissions() []Submission {
  submissions := make([]Submission, 0)
  var submission Submission
  iter := Submissions.Find(nil).Sort("-receivedat").Iter()
  for iter.Next(&submission) {
    submissions = append(submissions, submission)
  }
  return submissions
}

/* Main queue, should be fast because everyone is slamming this page */
func SubmissionsIndex(w http.ResponseWriter, r *http.Request) {
  check(queuet.Execute(w, AllSubmissions()))
}

/* Main solution progress scoreboard */
func ProgressIndex(w http.ResponseWriter, r *http.Request) {
  data := struct{
    Teams []Team
    Solutions SolutionList
    Puzzles []Puzzle
  } {AllTeams(), AllSolutions(), AllPuzzles()}
  check(solutionst.Execute(w, data))
}

func SolutionFor(l SolutionList, t *Team, p *Puzzle) *Solution {
  for i, s := range l {
    if s.TeamId == t.Id && s.PuzzleId == p.Id {
      return &l[i]
    }
  }
  return nil
}

func NumSolved(l SolutionList, t *Team) int {
  cnt := 0
  for _, s := range l {
    if s.TeamId == t.Id && s.SolvedAt.Year() > 1000 {
      cnt++
    }
  }
  return cnt
}

func ProgressReset(w http.ResponseWriter, r *http.Request) {
  _, err := Solutions.RemoveAll(nil)
  check(err)
  _, err = Submissions.RemoveAll(nil)
  check(err)
  http.Redirect(w, r, "/admin/progress", http.StatusFound)
}

func ProgressRelease(w http.ResponseWriter, r *http.Request) {
  /* Delete all previous solutions/submissions (also redirect) */
  ProgressReset(w, r)
  var puzzle Puzzle
  teams := AllTeams()
  iter := Puzzles.Find(bson.M{"secondround":false,
                              "metapuzzle":false,
                              "unlockidx":bson.M{"$lte": 0}}).Iter()

  for iter.Next(&puzzle) {
    for i, _ := range teams {
      _, err := CreateSolution(&teams[i], &puzzle)
      check(err)
    }
  }
}

func SubmissionRespond(w http.ResponseWriter, r *http.Request) {
  var submission Submission
  var puzzle Puzzle
  var team Team
  var solution Solution
  submission.find(mux.Vars(r)["id"])
  solution.findId(submission.SolutionId)
  puzzle.findId(solution.PuzzleId)
  team.findId(solution.TeamId)
  submission.Respond(&puzzle, &team, r.FormValue("response"))
  submission.Status = IncorrectReplied
  check(Submissions.UpdateId(submission.Id, submission))

  http.Redirect(w, r, "/admin/queue", http.StatusFound)
}

func (s *Submission) Respond(p *Puzzle, t *Team, content string) {
  // TODO: send this response to the team
}

func (s *Submission) NeedsResponse() bool {
  return s.Status == IncorrectUnreplied
}

func (s *Submission) find(id string) {
  check(Submissions.FindId(bson.ObjectIdHex(id)).One(s))
}

func (s *Solution) findId(id bson.ObjectId) {
  check(Solutions.FindId(id).One(s))
}

func (s SubmissionStatus) String() string {
  switch (s) {
    case Correct, CorrectUnreplied: return "correct"
    case InvalidAnswer: return "invalid-answer"
    case IncorrectReplied: return "incorrect-replied"
    case IncorrectUnreplied: return "incorrect-unreplied"
  }

  return "invalid-status"
}

func (s *Submission) DisplayAnswer() string {
  if s.Status == InvalidAnswer {
    return "<invalid>"
  }
  return s.Answer
}

func CreateSolution(t *Team, p *Puzzle) (*Solution, error) {
  solution := Solution{TeamId: t.Id, PuzzleId: p.Id}
  err := Solutions.Insert(&solution)
  return &solution, err
}
