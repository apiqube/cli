package save

import (
	"fmt"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

type Result struct {
	ManifestID string
	CaseName   string
	Target     string
	Method     string

	ResultCase *interfaces.CaseResult

	Request  *Entry
	Response *Entry
}

type Entry struct {
	Headers map[string]string
	Body    map[string]any
}

func FormSaveKey(manifestID, caseName, suffix string) string {
	return fmt.Sprintf("%s.%s.%s.%s", KeyPrefix, manifestID, caseName, suffix)
}
