package main

import "time"

type DeadlineStore struct{ learners []learner }

func NewDeadlineStore(ls []learner) *DeadlineStore { return &DeadlineStore{learners: ls} }

type reportRow struct {
	Email  string `json:"email"`
	Course string `json:"course"`
	Status string `json:"status"`
}

func (s *DeadlineStore) Report(now time.Time) []reportRow {
	out := make([]reportRow, 0, len(s.learners))
	for _, l := range s.learners {
		status := "on_track"
		if now.After(l.Due) {
			status = "overdue"
		}
		out = append(out, reportRow{l.Email, l.Course, status})
	}
	return out
}
