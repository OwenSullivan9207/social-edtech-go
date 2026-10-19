package main

import (
	"testing"
	"time"
)

func TestReportMarksDeadline(t *testing.T) {
	due := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	s := NewDeadlineStore([]learner{{Email: "learner@example.edu", Course: "ETL", Due: due}})
	got := s.Report(due.Add(time.Minute))
	if got[0].Status != "overdue" {
		t.Fatalf("status=%q", got[0].Status)
	}
}
