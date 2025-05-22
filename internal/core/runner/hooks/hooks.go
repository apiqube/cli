package hooks

import (
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

const hooksRunnerOutputPrefix = "Hooks Runner:"

type Runner interface {
	RunHooks(ctx interfaces.ExecutionContext, event HookEvent, actions []Action) error
	RegisterHooksHandler(event HookEvent, handler HookHandler)
}

type HookEvent string

func (h HookEvent) String() string {
	return string(h)
}

const (
	BeforeRun HookEvent = "before run"
	AfterRun  HookEvent = "after run"
	OnSuccess HookEvent = "on success"
	OnFailure HookEvent = "on failure"
)

type HookHandler func(ctx interfaces.ExecutionContext, actions []Action) error

type Action struct {
	Type   string         `yaml:"type" json:"type" validate:"required,oneof=log save skip fail exec notify"` // eg log/save/skip/fail/exec/notify
	Params map[string]any `yaml:"params" json:"params" validate:"required"`
}

type DefaultHooksRunner struct {
	entries map[HookEvent][]HookHandler
}

func NewDefaultHooksRunner() *DefaultHooksRunner {
	return &DefaultHooksRunner{
		entries: make(map[HookEvent][]HookHandler),
	}
}

func (r *DefaultHooksRunner) RunHooks(ctx interfaces.ExecutionContext, event HookEvent, actions []Action) error {
	if len(actions) == 0 {
		return nil
	}

	switch event {
	case BeforeRun:
		return r.runBeforeRunHooks(ctx, event, actions)
	case AfterRun:
		return r.runAfterRunHooks(ctx, event, actions)
	case OnSuccess:
		return r.runOnSuccessHooks(ctx, event, actions)
	case OnFailure:
		return r.runOnFailureHooks(ctx, event, actions)
	default:
		return nil
	}
}

func (r *DefaultHooksRunner) RegisterHooksHandler(event HookEvent, handler HookHandler) {
	r.entries[event] = append(r.entries[event], handler)
}

func (r *DefaultHooksRunner) runBeforeRunHooks(ctx interfaces.ExecutionContext, event HookEvent, actions []Action) error {
	output := ctx.GetOutput()
	output.Logf(interfaces.InfoLevel, "%s running %s hooks", hooksRunnerOutputPrefix, event.String())

	return nil
}

func (r *DefaultHooksRunner) runAfterRunHooks(ctx interfaces.ExecutionContext, event HookEvent, actions []Action) error {
	output := ctx.GetOutput()
	output.Logf(interfaces.InfoLevel, "%s running %s hooks", hooksRunnerOutputPrefix, event.String())

	return nil
}

func (r *DefaultHooksRunner) runOnSuccessHooks(ctx interfaces.ExecutionContext, event HookEvent, actions []Action) error {
	output := ctx.GetOutput()
	output.Logf(interfaces.InfoLevel, "%s running %s hooks", hooksRunnerOutputPrefix, event.String())

	return nil
}

func (r *DefaultHooksRunner) runOnFailureHooks(ctx interfaces.ExecutionContext, event HookEvent, actions []Action) error {
	output := ctx.GetOutput()
	output.Logf(interfaces.InfoLevel, "%s running %s hooks", hooksRunnerOutputPrefix, event.String())

	return nil
}
