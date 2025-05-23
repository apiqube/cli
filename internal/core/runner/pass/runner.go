package pass

import (
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (p *Runner) Apply(_ interfaces.ExecutionContext, input string, _ []tests.Pass) string {
	// заменить плейсхолдеры в URL: {{.token}}, {{.user.id}}, и т.п.
	// + обрабатывать Pass.Map
	return input
	// return ReplaceWithStoreValues(ctx, input, passes)
}

func (p *Runner) ApplyBody(_ interfaces.ExecutionContext, body map[string]any, _ []tests.Pass) map[string]any {
	// аналогично — пройтись по body и заменить шаблоны
	return body
}

func (p *Runner) MapHeaders(_ctx interfaces.ExecutionContext, headers map[string]string, _ []tests.Pass) map[string]string {
	// заменить плейсхолдеры в заголовках
	return headers
}
