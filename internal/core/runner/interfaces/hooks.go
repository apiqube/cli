package interfaces

type HookRunner interface {
	RunHook(event HookEvent, ctx ExecutionContext, metadata map[string]any) error
	RegisterHookHandler(event HookEvent, handler HookHandler)
}

type HookEvent string

const (
	beforeRun   HookEvent = "beforeRun"
	afterRun    HookEvent = "afterRun"
	BeforeStage HookEvent = "beforeStage"
	AfterStage  HookEvent = "afterStage"
	OnSuccess   HookEvent = "onSuccess"
	OnFailure   HookEvent = "onFailure"
)

type HookHandler func(ctx ExecutionContext, metadata map[string]any) error
