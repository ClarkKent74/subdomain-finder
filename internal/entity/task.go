package entity

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

func (s Status) IsActive() bool {
	return s == StatusPending || s == StatusRunning
}

type Task struct {
	Domain    string
	Status    Status
	Results   []string
	Error     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
