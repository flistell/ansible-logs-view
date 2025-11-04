package app

import (
	"time"
)

// Task represents a single Ansible task entry
type Task struct {
	ID          int
	Description string
	StartTime   time.Time
	Status      string // "ok", "changed", "skipping", "failed"
	Host        string
	Path        string
	Diff        string // Diff information for the task
	RawText     string // Raw text of the entire task from the log file
}

// DiffSection represents a diff section in a task
type DiffSection struct {
	BeforeFile string
	AfterFile  string
	Content    string
}
