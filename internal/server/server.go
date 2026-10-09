// Package server serves the exam panel: the question list and timer you keep next to your
// terminal, and the grading that runs when you end the exam.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/internal/tasks"
	"github.com/MohamedAljoke/cka-sim/internal/terminal"
)

const PassScore = 66

// Exam is the persisted state of one sitting, so a restarted server resumes the same exam.
// A study session is an Exam without a clock, where each task can be checked, reset and
// explained on its own.
type Exam struct {
	Study     bool            `json:"study,omitempty"`
	StartedAt time.Time       `json:"startedAt"`
	Duration  int             `json:"durationSeconds"`
	TaskIDs   []string        `json:"taskIds"`
	Flags     map[string]bool `json:"flags"`
	EndedAt   *time.Time      `json:"endedAt,omitempty"`
	Results   []grader.Result `json:"results,omitempty"`
	Score     float64         `json:"score"`
}

func (e *Exam) Deadline() time.Time {
	return e.StartedAt.Add(time.Duration(e.Duration) * time.Second)
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(stateDir string) *Store { return &Store{path: filepath.Join(stateDir, "exam.json")} }

func (s *Store) Load() (*Exam, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	var e Exam
	return &e, json.Unmarshal(data, &e)
}

func (s *Store) Save(e *Exam) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

type Server struct {
	Store  *Store
	Tasks  []tasks.Task
	Runner grader.Runner
	UI     fs.FS
	// Terminals backs the browser terminal; without it the panel shows none.
	Terminals *terminal.Manager

	grading sync.Mutex
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/exam", s.getExam)
	mux.HandleFunc("POST /api/flag/{id}", s.toggleFlag)
	mux.HandleFunc("POST /api/end", s.endExam)
	mux.HandleFunc("POST /api/check/{id}", s.studyOnly(s.checkTask))
	mux.HandleFunc("POST /api/reset/{id}", s.studyOnly(s.resetTask))
	if s.Terminals != nil {
		mux.HandleFunc("GET /api/terminals", s.listTerminals)
		mux.HandleFunc("POST /api/terminals", s.startTerminal)
		mux.HandleFunc("DELETE /api/terminals/{id}", s.closeTerminal)
		mux.HandleFunc("GET /api/terminals/{id}/ws", func(w http.ResponseWriter, r *http.Request) {
			s.Terminals.Attach(w, r, r.PathValue("id"))
		})
	}
	mux.Handle("GET /", http.FileServerFS(s.UI))
	return sameOrigin(mux)
}

// sameOrigin refuses changes requested by another website's page: the panel only listens on
// localhost, but any page the browser opens can still send requests there.
func sameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if origin := r.Header.Get("Origin"); origin != "" {
				if u, err := url.Parse(origin); err != nil || u.Host != r.Host {
					http.Error(w, "cross-origin request refused", http.StatusForbidden)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) listTerminals(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.Terminals.List())
}

func (s *Server) startTerminal(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.Terminals.Start())
}

func (s *Server) closeTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.Terminals.Close(r.PathValue("id")) {
		http.Error(w, "no such terminal", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type questionView struct {
	Number   int            `json:"number"`
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Domain   string         `json:"domain"`
	Weight   int            `json:"weight"`
	Host     string         `json:"host"`
	Body     string         `json:"body"`
	Flagged  bool           `json:"flagged"`
	Result   *grader.Result `json:"result,omitempty"`
	Explain  string         `json:"explain,omitempty"`
	Solution string         `json:"solution,omitempty"`
	Fraction float64        `json:"fraction"`
}

type examView struct {
	Study     bool           `json:"study"`
	StartedAt time.Time      `json:"startedAt"`
	Deadline  time.Time      `json:"deadline"`
	Now       time.Time      `json:"now"`
	Ended     bool           `json:"ended"`
	Score     float64        `json:"score"`
	PassScore int            `json:"passScore"`
	Questions []questionView `json:"questions"`
}

func (s *Server) view(e *Exam) examView {
	v := examView{
		Study:     e.Study,
		StartedAt: e.StartedAt,
		Deadline:  e.Deadline(),
		Now:       time.Now(),
		Ended:     e.EndedAt != nil,
		Score:     e.Score,
		PassScore: PassScore,
	}
	results := map[string]grader.Result{}
	for _, r := range e.Results {
		results[r.TaskID] = r
	}
	for i, id := range e.TaskIDs {
		t, ok := tasks.Find(s.Tasks, id)
		if !ok {
			continue
		}
		q := questionView{
			Number:  i + 1,
			ID:      t.ID,
			Title:   t.Title,
			Domain:  tasks.DomainTitles[t.Domain],
			Weight:  t.Weight,
			Host:    t.Host,
			Body:    t.Body,
			Flagged: e.Flags[t.ID],
		}
		// In an exam, results, explanations and solutions stay hidden until it is over.
		if r, graded := results[id]; graded && (v.Ended || e.Study) {
			q.Result = &r
			q.Fraction = r.Fraction()
		}
		if (q.Result != nil && v.Ended) || e.Study {
			q.Explain = t.Explain
			if data, err := os.ReadFile(filepath.Join(s.Runner.AssetsDir, t.Dir, "solution.sh")); err == nil {
				q.Solution = string(data)
			}
		}
		v.Questions = append(v.Questions, q)
	}
	return v
}

func (s *Server) getExam(w http.ResponseWriter, _ *http.Request) {
	e, err := s.Store.Load()
	if err != nil {
		http.Error(w, "no exam in progress — start one with: cka-sim exam", http.StatusNotFound)
		return
	}
	writeJSON(w, s.view(e))
}

func (s *Server) toggleFlag(w http.ResponseWriter, r *http.Request) {
	e, err := s.Store.Load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if e.Flags == nil {
		e.Flags = map[string]bool{}
	}
	id := r.PathValue("id")
	e.Flags[id] = !e.Flags[id]
	if err := s.Store.Save(e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, s.view(e))
}

func (s *Server) endExam(w http.ResponseWriter, r *http.Request) {
	// One grading run at a time: a double click must not grade twice.
	s.grading.Lock()
	defer s.grading.Unlock()

	e, err := s.Store.Load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if e.EndedAt == nil {
		var exam []tasks.Task
		for _, id := range e.TaskIDs {
			if t, ok := tasks.Find(s.Tasks, id); ok {
				exam = append(exam, t)
			}
		}
		// Grading outlives a closed browser tab.
		e.Results = s.Runner.GradeAll(context.WithoutCancel(r.Context()), exam)
		e.Score = Score(exam, e.Results)
		now := time.Now()
		e.EndedAt = &now
		if err := s.Store.Save(e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, s.view(e))
}

// studyOnly guards the per-task actions: the real exam never tells you how a task is going.
func (s *Server) studyOnly(next func(http.ResponseWriter, *http.Request, *Exam, tasks.Task)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.grading.Lock()
		defer s.grading.Unlock()
		e, err := s.Store.Load()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if !e.Study {
			http.Error(w, "only available in study mode — start one with: cka-sim study", http.StatusForbidden)
			return
		}
		t, ok := tasks.Find(s.Tasks, r.PathValue("id"))
		if !ok || !slices.Contains(e.TaskIDs, t.ID) {
			http.Error(w, "no such task in this session", http.StatusNotFound)
			return
		}
		next(w, r, e, t)
	}
}

func (s *Server) checkTask(w http.ResponseWriter, r *http.Request, e *Exam, t tasks.Task) {
	result := s.Runner.Grade(context.WithoutCancel(r.Context()), t)
	e.Results = slices.DeleteFunc(e.Results, func(old grader.Result) bool { return old.TaskID == t.ID })
	e.Results = append(e.Results, result)
	s.save(w, e)
}

// resetTask runs the task's setup again, so you can retry it from its starting state.
func (s *Server) resetTask(w http.ResponseWriter, r *http.Request, e *Exam, t tasks.Task) {
	if out, err := s.Runner.Script(context.WithoutCancel(r.Context()), t, "setup.sh", 6*time.Minute); err != nil {
		http.Error(w, fmt.Sprintf("resetting %s: %v\n%s", t.ID, err, out), http.StatusInternalServerError)
		return
	}
	e.Results = slices.DeleteFunc(e.Results, func(old grader.Result) bool { return old.TaskID == t.ID })
	s.save(w, e)
}

func (s *Server) save(w http.ResponseWriter, e *Exam) {
	if err := s.Store.Save(e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, s.view(e))
}

// Score weighs each task's earned fraction by the task's exam weight, as a percentage.
func Score(exam []tasks.Task, results []grader.Result) float64 {
	byID := map[string]grader.Result{}
	for _, r := range results {
		byID[r.TaskID] = r
	}
	var earned, total float64
	for _, t := range exam {
		total += float64(t.Weight)
		earned += float64(t.Weight) * byID[t.ID].Fraction()
	}
	if total == 0 {
		return 0
	}
	return math.Round(earned/total*1000) / 10
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil && !errors.Is(err, http.ErrHandlerTimeout) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
