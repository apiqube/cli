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
