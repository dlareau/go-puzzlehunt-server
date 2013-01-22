package main

import "github.com/gorilla/mux"
import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"
import "net/http"
import "net/mail"
import "regexp"
import "strings"
import "time"
import uuid "github.com/nu7hatch/gouuid"

import "puzzlehunt/email"

type Solution struct {
  Id        bson.ObjectId "_id,omitempty"
  TeamId    bson.ObjectId
  PuzzleId  bson.ObjectId
  SolvedAt  time.Time
}

type Submission struct {
  Id          bson.ObjectId "_id,omitempty"
  SolutionId  bson.ObjectId
  Answer      string
  Status      SubmissionStatus
  ReceivedAt  time.Time
  MessageId   string
  References  string
  Subject     string
}

type SubmissionStatus int

const (
  Correct SubmissionStatus = iota
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

var emailRegex = regexp.MustCompile(`([^\+]*)(?:\+(.*))?@.*`)

func EmailReceived(w http.ResponseWriter, r *http.Request) {
  msg, err := mail.ReadMessage(strings.NewReader(r.FormValue("mail")))
  check(err)

  /* Parse the To: address to find out what puzzle is being submitted */
  to, err := msg.Header.AddressList("To")
  check(err)
  matches := emailRegex.FindStringSubmatch(to[0].Address)
  if len(matches) != 3 {
    panic("Bad email in To:")
  }
  var puzzle Puzzle
  err = Puzzles.Find(bson.M{"slug": matches[2]}).One(&puzzle)
  check(err)

  /* Parse the From: to figure out who submitted the puzzle */
  from, err := msg.Header.AddressList("From")
  check(err)
  matches = emailRegex.FindStringSubmatch(from[0].Address)
  if len(matches) != 3 {
    panic("Bad email in From:")
  }
  var team Team
  err = Teams.Find(bson.M{"emailaddress": from[0].Address}).One(&team)
  check(err)

  /* Find the solution struct for this team/puzzle pair */
  var solution Solution
  err = Solutions.Find(bson.M{"teamid": team.Id,
                              "puzzleid": puzzle.Id}).One(&solution)
  check(err)

  /* Create the submission */
  var submission Submission
  submission.SolutionId = solution.Id
  date, err := msg.Header.Date()
  check(err)
  submission.ReceivedAt = date
  submission.MessageId = msg.Header.Get("Message-Id")
  submission.Subject = msg.Header.Get("Subject")
  submission.References = msg.Header.Get("References")
  defer Submissions.Insert(&submission)
  defer w.WriteHeader(http.StatusOK)

  /* Figure out the answer they gave us */
  data, err := email.Plaintext(msg)
  check(err)
  submission.Answer = strings.TrimSpace(string(data))
  if strings.Index(submission.Answer, " ") > 0 {
    submission.Respond(&puzzle, &team, InvalidEmailFormat)
    submission.Status = InvalidAnswer
    return
  }

  /* If the answer is right, email back that it's right */
  if strings.EqualFold(submission.Answer, puzzle.Answer) {
    solution.SolvedAt = submission.ReceivedAt
    err = Solutions.UpdateId(solution.Id, &solution)
    check(err)
    hasmore := false
    if !puzzle.SecondRound {
      hasmore, err = team.UnlockMore()
      check(err)
    }
    submission.Status = Correct
    text := EmailCorrectAnswer
    if hasmore {
      text = EmailCorrectMorePuzzles
    }
    submission.Respond(&puzzle, &team, text)
  } else {
    /* Nothing to do, it's in the queue and it'll be responded to soon */
    submission.Status = IncorrectUnreplied
  }
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
  for _, t := range teams {
    var msg email.Message
    msg.From = mail.Address{Address: Round1Username + "@" + EmailHost,
                            Name: Round1Name}
    msg.To = []mail.Address{t.Email()}
    msg.Subject = EmailInitialSubject
    msg.Content = EmailInitialBody

    msg.Headers = map[string]string{"Content-Type": "text/plain; charset=utf8"}
    check(msg.Send(MailServer))
  }

  for iter.Next(&puzzle) {
    for i, _ := range teams {
      _, err := CreateSolution(&teams[i], &puzzle)
      check(err)
    }
  }
}

func SubmissionsIndex(w http.ResponseWriter, r *http.Request) {
  data := struct{
    Teams       map[string]Team
    Solutions   map[string]Solution
    Puzzles     map[string]Puzzle
    Submissions []Submission
  }{make(map[string]Team), make(map[string]Solution), make(map[string]Puzzle),
    AllSubmissions()}

  all := func(c *mgo.Collection, ids []bson.ObjectId) *mgo.Iter {
    return c.Find(bson.M{"_id": bson.M{"$in": ids}}).Iter()
  }

  /* Find all Solutions */
  ids := []bson.ObjectId{}
  for _, submission := range data.Submissions {
    ids = append(ids, submission.SolutionId)
  }
  var soln Solution
  puzzle_ids := []bson.ObjectId{}
  team_ids := []bson.ObjectId{}
  for iter := all(Solutions, ids); iter.Next(&soln); {
    data.Solutions[soln.Id.Hex()] = soln
    puzzle_ids = append(puzzle_ids, soln.PuzzleId)
    team_ids = append(team_ids, soln.TeamId)
  }

  /* Now find all puzzles/teams */
  var puzzle Puzzle
  var team Team
  for iter := all(Teams, team_ids); iter.Next(&team); {
    data.Teams[team.Id.Hex()] = team
  }
  for iter := all(Puzzles, puzzle_ids); iter.Next(&puzzle); {
    data.Puzzles[puzzle.Id.Hex()] = puzzle
  }

  check(queuet.Execute(w, data))
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
  uid, _ := uuid.NewV4()
  id := "<" + uid.String() + "@" + EmailHost + ">"
  if s.Status == IncorrectUnreplied || s.Status == IncorrectReplied ||
     s.Status == Correct {
    content = "> " + s.Answer + "\n\n" + content
  }

  msg := email.Message{
    From: p.FromAddress(),
    To: []mail.Address{t.Email()},
    Subject: s.Subject,
    Content: content,
    Headers: map[string]string{"In-Reply-To": s.MessageId,
                               "References": s.MessageId + " " + s.References,
                               "Message-Id": id},
  }
  check(msg.Send(MailServer))
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
    case Correct: return "correct"
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
  if err == nil {
    err = p.EmailTo(t)
  }
  return &solution, err
}
