package main

import "github.com/gorilla/mux"
import "labix.org/v2/mgo/bson"
import "net/http"
import "strings"
import "sync"
import "time"

var CorrectNotifiers sync.WaitGroup

func MapHandler(w http.ResponseWriter, r *http.Request, t *Team) {
	all := AllPuzzles()
	var soln Solution
	solns := make([]Solution, 0)
	for iter := Solutions.Find(bson.M{"teamid": t.Id}).Iter(); iter.Next(&soln); {
		solns = append(solns, soln)
	}
	puzzles := make([]Puzzle, 0)
	for _, puz := range all {
		for _, soln := range solns {
			if puz.Id == soln.PuzzleId {
				puzzles = append(puzzles, puz)
				break
			}
		}
	}
	data := struct {
		Solutions SolutionList
		Team      *Team
		Puzzles   []Puzzle
	}{solns, t, AllPuzzles()}

	check(Template("_base.html", "puzzles.html").Execute(w, data))
}

func MapPuzzleHandler(w http.ResponseWriter, r *http.Request, t *Team) {
	var puzzle Puzzle
	puzzle.findSlug(mux.Vars(r)["id"])
	var soln Solution
	check(Solutions.Find(bson.M{"teamid": t.Id, "puzzleid": puzzle.Id}).One(&soln))

	if r.Method == "POST" {
		answer := r.FormValue("answer")
		submission := &Submission{SolutionId: soln.Id,
			TeamName:   t.Name,
			PuzzleName: puzzle.Name,
			Answer:     answer,
			ReceivedAt: time.Now()}
		if strings.EqualFold(answer, puzzle.Answer) {
			submission.Status = CorrectUnreplied
			go updateCorrect(submission, &soln)
		} else if strings.Index(answer, " ") != -1 {
			submission.Status = InvalidAnswer
		} else {
			submission.Status = IncorrectUnreplied
		}
		check(submission.Insert())
	}

	submissions := make([]Submission, 0)
	var s Submission
	iter := Submissions.Find(bson.M{"solutionid": soln.Id}).
		Sort("-receivedat").Iter()
	for iter.Next(&s) {
		submissions = append(submissions, s)
	}

	data := struct {
		Puzzle      *Puzzle
		Submissions []Submission
		Tag         string
	}{&puzzle, submissions, t.Id.Hex() + puzzle.Id.Hex()}

	check(Template("_base.html", "puzzle.html").Execute(w, &data))
}

func updateCorrect(s *Submission, soln *Solution) {
	CorrectNotifiers.Add(1)
	time.Sleep(5 * time.Second)
	s.Status = Correct
	soln.SolvedAt = time.Now()
	check(s.Update())

	// todo: check this. why is erring but working?
	soln.Update()
	CorrectNotifiers.Add(-1)
}

func ChartsPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Teams       []Team
		Puzzles     []Puzzle
		Solutions   []Solution
		Submissions []Submission
	}{AllTeams(), AllPuzzles(), AllSolutions(), AllSubmissions()}
	check(AdminTemplate("charts.html").Execute(w, data))
}
