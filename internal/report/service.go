package report

import (
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/save"
)

// Service aggregates results from ExecutionContext and generates reports.
type Service struct {
	generator Generator
}

// NewReportService creates a new ReportService with the given generator.
func NewReportService(generator Generator) *Service {
	return &Service{generator: generator}
}

// CollectResults collects all save.Result from the context by manifest IDs.
func (s *Service) CollectResults(ctx interfaces.ExecutionContext) []*save.Result {
	mans := ctx.GetAllManifests()
	results := make([]*save.Result, 0, len(mans))

	for _, man := range mans {
		key := save.FormSaveKey(man.GetID(), save.ResultKeySuffix)
		if val, ok := ctx.Get(key); ok {
			if res, is := val.([]*save.Result); is {
				results = append(results, res...)
			}
		}
	}

	return results
}

func (s *Service) GenerateReports(ctx interfaces.ExecutionContext) error {
	return s.generator.Generate(s.CollectResults(ctx))
}
