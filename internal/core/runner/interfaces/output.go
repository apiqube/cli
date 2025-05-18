package interfaces

import (
	"github.com/apiqube/cli/internal/core/manifests"
)

type Output interface {
	StartCase(manifest manifests.Manifest, caseName string)
	EndCase(manifest manifests.Manifest, caseName string, result *CaseResult)
	ReceiveMsg(msg any)
	Log(level LogLevel, msg string)
	Logf(level LogLevel, format string, args ...any)
	DumpValues(values map[string]any)
	Error(err error)
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
