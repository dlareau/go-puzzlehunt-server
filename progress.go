package main

import "bytes"
import "github.com/gorilla/mux"
import "labix.org/v2/mgo"
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

type QueueMessage struct {
  Html string
  Id   string
  Type string
}

type ProgressMessage struct {
  Html string
  Id   string
}

var Solutions = db.C("solutions")
var Submissions = db.C("submissions")

var solutionst = AdminTemplate("progress/solutions.html")
var queuet = AdminTemplate("progress/queue.html")
var Queue = WsServer()
var Progress = WsServer()
var PuzzleStatus = TagWsServer(func(r *http.Request) string {
  return mux.Vars(r)["tag"]
})

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

func PuzzleSolved(l SolutionList, t *Team, p *Puzzle) bool {
  for i, s := range l {
    if s.TeamId == t.Id && s.PuzzleId == p.Id {
      return !l[i].SolvedAt.IsZero()
    }
  }
  return false
}

func SolutionFor(l SolutionList, t *Team, p *Puzzle) *Solution {
  for i, s := range l {
    if s.TeamId == t.Id && s.PuzzleId == p.Id {
      return &l[i]
    }
  }
  return nil
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
  iter := Puzzles.Find(bson.M{"metapuzzle":false,
                              "unlockidx":bson.M{"$lte": 5}}).Iter()

  for iter.Next(&puzzle) {
    for i, _ := range teams {
      solution := Solution{TeamId: teams[i].Id, PuzzleId: puzzle.Id}
      check(solution.Insert())
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
  submission.Comment = r.FormValue("response")
  submission.Status = IncorrectReplied
  check(submission.Update())
}

func (s *Submission) NeedsResponse() bool {
  return s.Status == IncorrectUnreplied
}

func (s *Submission) find(id string) {
  check(Submissions.FindId(bson.ObjectIdHex(id)).One(s))
}

func (s *Submission) qmessage(typ string) QueueMessage {
  buf := bytes.NewBuffer(make([]byte, 0))
  check(queuet.ExecuteTemplate(buf, "queue_submission", s))
  return QueueMessage{ Id: s.Id.Hex(), Html: buf.String(), Type: typ }
}

func (s *Submission) pmessage(typ string) QueueMessage {
  buf := bytes.NewBuffer(make([]byte, 0))
  check(mpuzzle.ExecuteTemplate(buf, "submission", s))
  return QueueMessage{ Id: s.Id.Hex(), Html: buf.String(), Type: typ }
}

func (s *Submission) Insert() error {
  s.Id = bson.NewObjectId()
  err := Submissions.Insert(s)
  if err == nil {
    Queue.Broadcast <- s.qmessage("new")
    PuzzleStatus.Tags <- TaggedMessage{ Tag: s.Tag(), Msg: s.pmessage("new") }
  }
  return err
}

func (s *Submission) Update() error {
  err := Submissions.UpdateId(s.Id, s)
  if err == nil {
    Queue.Broadcast <- s.qmessage("update")
    PuzzleStatus.Tags <- TaggedMessage{ Tag: s.Tag(), Msg: s.pmessage("update") }
  }
  return err
}

func (s *Submission) DisplayAnswer() string {
  if s.Status == InvalidAnswer {
    return "<invalid>"
  }
  return s.Answer
}

func (s *Submission) AnswerStatus() string {
  switch (s.Status) {
    case CorrectUnreplied, IncorrectUnreplied: return "validating...";
    case Correct:                              return "correct";
    case InvalidAnswer:                        return InvalidAnswerText;
    case IncorrectReplied:                     return "incorrect: " + s.Comment;
  }
  return ""
}

func (s *Submission) Tag() string {
  var soln Solution
  soln.findId(s.SolutionId)
  return soln.TeamId.Hex() + soln.PuzzleId.Hex()
}

func (s *Solution) message() ProgressMessage {
  buf := bytes.NewBuffer(make([]byte, 0))
  check(solutionst.ExecuteTemplate(buf, "solution", s))
  return ProgressMessage{ Id: s.Identifier(),
                          Html: buf.String() }
}

func (s *Solution) findId(id bson.ObjectId) {
  check(Solutions.FindId(id).One(s))
}

func (s *Solution) Insert() error {
  err := Solutions.Insert(s)
  if err == nil {
    Progress.Broadcast <- s.message()
  }
  return err
}

func (s *Solution) Update() error {
  err := Solutions.UpdateId(s.Id, s)
  if err == nil {
    Progress.Broadcast <- s.message()

    var puzzle Puzzle
    puzzle.findId(s.PuzzleId)
    to_unlock := UnlockTree[puzzle.UnlockIdx]

    /* For everything we're supposed to unlock, see if it's already unlocked and
       if it isn't, insert the Solution to indicate that it's now available for
       solving */
    for _, idx := range to_unlock {
      /* find the puzzle to unlock */
      err = Puzzles.Find(bson.M{"unlockidx":idx}).One(&puzzle)
      if err != nil {
        return err
      }
      /* if it's already unlocked, no need to unlock again */
      n, err := Solutions.Find(bson.M{"puzzleid":puzzle.Id, "teamid":s.TeamId}).
                          Count()
      if err != nil || n == 1 {
        return err
      }

      /* If we're unlocking the meta, we require more solutions than 1 */
      if len(to_unlock) == 1 && to_unlock[0] == MetaIndex {
        iter := Puzzles.Find(bson.M{"unlockidx":bson.M{"$gte":MetaMinimum}}).Iter()
        var p Puzzle
        solved := 0
        for iter.Next(&p) {
          var soln Solution
          err := Solutions.Find(bson.M{"puzzleid":p.Id,
                                       "teamid":s.TeamId}).One(&soln)
          if err == mgo.ErrNotFound {
            continue
          } else if err != nil {
            return err
          } else if soln.SolvedAt.Year() > 1400 {
            solved += 1
          }
        }

        if solved < MetaRequired {
          return nil
        }
      }

      /* Finally, actually unlock the puzzle */
      solution := Solution{TeamId: s.TeamId, PuzzleId: puzzle.Id}
      err = solution.Insert()
      if err != nil {
        return err
      }
    }
  }

  return err
}

func (s *Solution) Identifier() string {
  return s.PuzzleId.Hex() + s.TeamId.Hex()
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
