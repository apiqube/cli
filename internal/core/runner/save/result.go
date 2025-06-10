package save

import (
	"fmt"
	"time"
)

type Result struct {
	ManifestID string
	CaseName   string
	Target     string
	Method     string
	Duration   time.Duration
	StatusCode int

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
