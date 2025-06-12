package report

import "github.com/apiqube/cli/internal/core/runner/save"

type Generator interface {
	Generate(results []*save.Result) error
}
