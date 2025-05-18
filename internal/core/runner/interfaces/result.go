package interfaces

import (
	"time"
)

type CaseResult struct {
	Name       string
	Success    bool
	StatusCode int
	Duration   time.Duration
	Errors     []string
	Values     map[string]any
	Details    map[string]any
}
