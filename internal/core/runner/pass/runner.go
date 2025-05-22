package pass

import (
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (p *Runner) Apply(ctx interfaces.ExecutionContext, input string, passes []tests.Pass) string {
	// заменить плейсхолдеры в URL: {{.token}}, {{.user.id}}, и т.п.
	// + обрабатывать Pass.Map
	return ""
	// return ReplaceWithStoreValues(ctx, input, passes)
}

func (p *Runner) ApplyBody(ctx interfaces.ExecutionContext, body map[string]any, passes []tests.Pass) map[string]any {
	// аналогично — пройтись по body и заменить шаблоны
	return body
}

func (p *Runner) MapHeaders(ctx interfaces.ExecutionContext, headers map[string]string, passes []tests.Pass) map[string]string {
	// заменить плейсхолдеры в заголовках
	return headers
}
