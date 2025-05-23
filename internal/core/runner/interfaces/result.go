package interfaces

import (
	"time"
)

type CaseResult struct {
	Name       string
	Success    bool
	Assert     string
	StatusCode int
	Duration   time.Duration
	Errors     []string
	Values     map[string]any
	Details    map[string]any
}

type Message struct {
	Format string
	Values []any
}

type Progress struct {
	Stage string
	Step  string
	Total int
	Done  int
}

type MetricResult struct {
	Name  string
	Value float64
	Unit  string
	Warn  float64
	Error float64
}
