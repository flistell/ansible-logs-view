package main

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
}