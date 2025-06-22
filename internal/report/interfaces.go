package report

import (
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

type Generator interface {
	Generate(ctx interfaces.ExecutionContext) error
}
