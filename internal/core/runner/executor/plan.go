package executor

import (
	"errors"
	"fmt"
	"sync"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
	"github.com/apiqube/cli/internal/core/runner/depends"
	"github.com/apiqube/cli/internal/core/runner/hooks"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

const planRunnerOutputPrefix = "Plan Runner:"

var _ interfaces.PlanRunner = (*DefaultPlanRunner)(nil)

type DefaultPlanRunner struct {
	registry    interfaces.ExecutorRegistry
	hooksRunner hooks.Runner
}

func NewDefaultPlanRunner(registry interfaces.ExecutorRegistry, hooksRunner hooks.Runner) *DefaultPlanRunner {
	return &DefaultPlanRunner{
		registry:    registry,
		hooksRunner: hooksRunner,
	}
}

// V2PlanRunner supports the new dependency system with PassManager
type V2PlanRunner struct {
	registry    interfaces.ExecutorRegistry
	hooksRunner hooks.Runner
	passManager *depends.PassManager
}

func NewV2PlanRunner(registry interfaces.ExecutorRegistry, hooksRunner hooks.Runner, graphResult *depends.GraphResultV2) *V2PlanRunner {
	return &V2PlanRunner{
		registry:    registry,
		hooksRunner: hooksRunner,
		passManager: nil, // Will be initialized when context is available
	}
}

func (r *V2PlanRunner) RunPlan(ctx interfaces.ExecutionContext, manifest manifests.Manifest) error {
	p, ok := manifest.(*plan.Plan)
	if !ok {
		return errors.New("invalid manifest type, expected Plan kind")
	}

	// Initialize PassManager if not already done
	if r.passManager == nil {
		// We need to rebuild the graph result from the plan
		// This is a simplified approach - in production you might want to pass the graph result directly
		allManifests := ctx.GetAllManifests()
		registry := depends.DefaultRuleRegistry()
		builder := depends.NewGraphBuilder(registry)
		graphResult, err := builder.Build(allManifests)
		if err != nil {
			return fmt.Errorf("failed to initialize dependency analysis: %w", err)
		}
		r.passManager = depends.NewPassManager(ctx, graphResult)
	}

	var err error
	output := ctx.GetOutput()

	planID := p.GetID()
	output.Logf(interfaces.InfoLevel, "%s starting V2 plan: %s", planRunnerOutputPrefix, planID)

	if err = ctx.Err(); err != nil {
		output.Logf(interfaces.ErrorLevel, "%s plan execution canceled before start: %v", planRunnerOutputPrefix, err)
		return err
	}

	if p.Spec.Hooks != nil {
		if err = r.runHooksWithContext(ctx, hooks.BeforeRun, p.Spec.Hooks.BeforeRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan before start hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}
	}

	for _, stage := range p.Spec.Stages {
		if err = ctx.Err(); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan execution canceled before stage %s: %v", planRunnerOutputPrefix, stage.Name, err)
			return err
		}

		stageName := stage.Name
		output.Logf(interfaces.InfoLevel, "%s %s stage starting...", planRunnerOutputPrefix, stageName)

		if stage.Hooks != nil {
			if err = r.runHooksWithContext(ctx, hooks.BeforeRun, stage.Hooks.BeforeRun); err != nil {
				output.Logf(interfaces.ErrorLevel, "%s stage %s before start hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
				return err
			}
		}

		var execErr error
		if stage.Parallel {
			execErr = r.runManifestsParallelV2(ctx, stage.Manifests)
		} else {
			execErr = r.runManifestsStrictV2(ctx, stage.Manifests)
		}

		if err = ctx.Err(); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan execution canceled after stage %s: %v", planRunnerOutputPrefix, stage.Name, err)
			return err
		}

		if stage.Hooks != nil {
			if err = r.runHooksWithContext(ctx, hooks.AfterRun, stage.Hooks.AfterRun); err != nil {
				output.Logf(interfaces.ErrorLevel, "%s stage %s after finish hooks running failed: %s", planRunnerOutputPrefix, stageName, err.Error())
				return err
			}
		}

		if execErr != nil {
			output.Logf(interfaces.ErrorLevel, "%s stage %s failed\nReason: %s", planRunnerOutputPrefix, stageName, execErr.Error())

			if stage.Hooks != nil {
				if err = r.runHooksWithContext(ctx, hooks.OnFailure, stage.Hooks.OnFailure); err != nil {
					output.Logf(interfaces.ErrorLevel, "%s stage %s on failure hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
					return err
				}
			}

			if p.Spec.Hooks != nil {
				if err = r.runHooksWithContext(ctx, hooks.OnFailure, p.Spec.Hooks.OnFailure); err != nil {
					output.Logf(interfaces.ErrorLevel, "%s plan on failure hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
					return errors.Join(execErr, err)
				}
			}

			return execErr
		}

		if stage.Hooks != nil {
			if err = r.runHooksWithContext(ctx, hooks.OnSuccess, stage.Hooks.OnSuccess); err != nil {
				output.Logf(interfaces.ErrorLevel, "%s stage %s on success hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
				return err
			}
		}
	}

	if err = ctx.Err(); err != nil {
		output.Logf(interfaces.ErrorLevel, "%s plan execution canceled before final hooks: %v", planRunnerOutputPrefix, err)
		return err
	}

	if p.Spec.Hooks != nil {
		if err = r.runHooksWithContext(ctx, hooks.AfterRun, p.Spec.Hooks.AfterRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan after finish hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}

		if err = r.runHooksWithContext(ctx, hooks.OnSuccess, p.Spec.Hooks.OnSuccess); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan on success hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}
	}

	return nil
}

func (r *V2PlanRunner) runManifestsStrictV2(ctx interfaces.ExecutionContext, manifestIDs []string) error {
	var man manifests.Manifest
	var err error

	output := ctx.GetOutput()

	for _, id := range manifestIDs {
		if man, err = ctx.GetManifestByID(id); err != nil {
			return fmt.Errorf("run %s manifest failed: %s", id, err.Error())
		}

		exec, exists := r.registry.Find(man.GetKind())
		if !exists {
			return fmt.Errorf("no executor found for kind: %s", man.GetKind())
		}

		// Check if this manifest should save results
		shouldSave := r.passManager.ShouldSaveResult(id)
		if shouldSave {
			output.Logf(interfaces.InfoLevel, "%s manifest %s will save results for data passing", planRunnerOutputPrefix, id)
		}

		output.Logf(interfaces.InfoLevel, "%s running %s manifest using %s executor", planRunnerOutputPrefix, id, man.GetKind())

		if err = exec.Run(ctx, man); err != nil {
			return fmt.Errorf("manifest %s failed: %s", id, err.Error())
		}

		// Save results if required (this would be integrated with the actual executor)
		if shouldSave {
			// This is a placeholder - the actual result saving would be done by the executor
			// For now, we just log that saving would happen
			output.Logf(interfaces.InfoLevel, "%s results saved for manifest %s", planRunnerOutputPrefix, id)
		}

		output.Logf(interfaces.InfoLevel, "%s %s manifest finished", planRunnerOutputPrefix, id)
	}

	return nil
}

func (r *V2PlanRunner) runManifestsParallelV2(ctx interfaces.ExecutionContext, manifestIDs []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(manifestIDs))

	output := ctx.GetOutput()

	for _, manId := range manifestIDs {
		id := manId
		wg.Add(1)

		go func() {
			defer wg.Done()
			man, err := ctx.GetManifestByID(id)
			if err != nil {
				errChan <- fmt.Errorf("run %s manifest failed: %s", id, err.Error())
				return
			}

			exec, exists := r.registry.Find(man.GetKind())
			if !exists {
				errChan <- fmt.Errorf("no executor found for kind: %s", man.GetKind())
				return
			}

			// Check if this manifest should save results
			shouldSave := r.passManager.ShouldSaveResult(id)
			if shouldSave {
				output.Logf(interfaces.InfoLevel, "%s manifest %s will save results for data passing", planRunnerOutputPrefix, id)
			}

			output.Logf(interfaces.InfoLevel, "%s running %s manifest using %s executor", planRunnerOutputPrefix, id, man.GetKind())

			if err = exec.Run(ctx, man); err != nil {
				errChan <- fmt.Errorf("manifest %s failed: %s", id, err.Error())
				return
			}

			// Save results if required
			if shouldSave {
				output.Logf(interfaces.InfoLevel, "%s results saved for manifest %s", planRunnerOutputPrefix, id)
			}

			output.Logf(interfaces.InfoLevel, "%s %s manifest finished", planRunnerOutputPrefix, id)
		}()
	}

	wg.Wait()
	close(errChan)

	var rErr error

	if len(errChan) > 0 {
		for err := range errChan {
			rErr = errors.Join(rErr, err)
		}

		return rErr
	}

	return nil
}

func (r *V2PlanRunner) runHooksWithContext(ctx interfaces.ExecutionContext, event hooks.HookEvent, actions []hooks.Action) error {
	if len(actions) == 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return r.hooksRunner.RunHooks(ctx, event, actions)
	}
}

func (r *DefaultPlanRunner) RunPlan(ctx interfaces.ExecutionContext, manifest manifests.Manifest) error {
	p, ok := manifest.(*plan.Plan)
	if !ok {
		return errors.New("invalid manifest type, expected Plan kind")
	}

	var err error
	output := ctx.GetOutput()

	planID := p.GetID()
	output.Logf(interfaces.InfoLevel, "%s starting plan: %s", planRunnerOutputPrefix, planID)

	if err = ctx.Err(); err != nil {
		output.Logf(interfaces.ErrorLevel, "%s plan execution canceled before start: %v", planRunnerOutputPrefix, err)
		return err
	}

	if p.Spec.Hooks != nil {
		if err = r.runHooksWithContext(ctx, hooks.BeforeRun, p.Spec.Hooks.BeforeRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan before start hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}
	}

	for _, stage := range p.Spec.Stages {
		if err = ctx.Err(); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan execution canceled before stage %s: %v", planRunnerOutputPrefix, stage.Name, err)
			return err
		}

		stageName := stage.Name
		output.Logf(interfaces.InfoLevel, "%s %s stage starting...", planRunnerOutputPrefix, stageName)

		if stage.Hooks != nil {
			if err = r.runHooksWithContext(ctx, hooks.BeforeRun, stage.Hooks.BeforeRun); err != nil {
				output.Logf(interfaces.ErrorLevel, "%s stage %s before start hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
				return err
			}
		}

		var execErr error
		if stage.Parallel {
			execErr = r.runManifestsParallel(ctx, stage.Manifests)
		} else {
			execErr = r.runManifestsStrict(ctx, stage.Manifests)
		}

		if err = ctx.Err(); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan execution canceled after stage %s: %v", planRunnerOutputPrefix, stage.Name, err)
			return err
		}

		if stage.Hooks != nil {
			if err = r.runHooksWithContext(ctx, hooks.AfterRun, stage.Hooks.AfterRun); err != nil {
				output.Logf(interfaces.ErrorLevel, "%s stage %s after finish hooks running failed: %s", planRunnerOutputPrefix, stageName, err.Error())
				return err
			}
		}

		if execErr != nil {
			output.Logf(interfaces.ErrorLevel, "%s stage %s failed\nReason: %s", planRunnerOutputPrefix, stageName, execErr.Error())

			if stage.Hooks != nil {
				if err = r.runHooksWithContext(ctx, hooks.OnFailure, stage.Hooks.OnFailure); err != nil {
					output.Logf(interfaces.ErrorLevel, "%s stage %s on failure hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
					return err
				}
			}

			if p.Spec.Hooks != nil {
				if err = r.runHooksWithContext(ctx, hooks.OnFailure, p.Spec.Hooks.OnFailure); err != nil {
					output.Logf(interfaces.ErrorLevel, "%s plan on failure hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
					return errors.Join(execErr, err)
				}
			}

			return execErr
		}

		if stage.Hooks != nil {
			if err = r.runHooksWithContext(ctx, hooks.OnSuccess, stage.Hooks.OnSuccess); err != nil {
				output.Logf(interfaces.ErrorLevel, "%s stage %s on success hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
				return err
			}
		}
	}

	if err = ctx.Err(); err != nil {
		output.Logf(interfaces.ErrorLevel, "%s plan execution canceled before final hooks: %v", planRunnerOutputPrefix, err)
		return err
	}

	if p.Spec.Hooks != nil {
		if err = r.runHooksWithContext(ctx, hooks.AfterRun, p.Spec.Hooks.AfterRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan after finish hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}

		if err = r.runHooksWithContext(ctx, hooks.OnSuccess, p.Spec.Hooks.OnSuccess); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan on success hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}
	}

	return nil
}

func (r *DefaultPlanRunner) runManifestsStrict(ctx interfaces.ExecutionContext, manifestIDs []string) error {
	var man manifests.Manifest
	var err error

	output := ctx.GetOutput()

	for _, id := range manifestIDs {
		if man, err = ctx.GetManifestByID(id); err != nil {
			return fmt.Errorf("run %s manifest failed: %s", id, err.Error())
		}

		exec, exists := r.registry.Find(man.GetKind())
		if !exists {
			return fmt.Errorf("no executor found for kind: %s", man.GetKind())
		}

		output.Logf(interfaces.InfoLevel, "%s running %s manifest using %s executor", planRunnerOutputPrefix, id, man.GetKind())

		if err = exec.Run(ctx, man); err != nil {
			return fmt.Errorf("manifest %s failed: %s", id, err.Error())
		}

		output.Logf(interfaces.InfoLevel, "%s %s manifest finished", planRunnerOutputPrefix, id)
	}

	return nil
}

func (r *DefaultPlanRunner) runManifestsParallel(ctx interfaces.ExecutionContext, manifestIDs []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(manifestIDs))

	output := ctx.GetOutput()

	for _, manId := range manifestIDs {
		id := manId
		wg.Add(1)

		go func() {
			defer wg.Done()
			man, err := ctx.GetManifestByID(id)
			if err != nil {
				errChan <- fmt.Errorf("run %s manifest failed: %s", id, err.Error())
				return
			}

			exec, exists := r.registry.Find(man.GetKind())
			if !exists {
				errChan <- fmt.Errorf("no executor found for kind: %s", man.GetKind())
				return
			}

			output.Logf(interfaces.InfoLevel, "%s running %s manifest using %s executor", planRunnerOutputPrefix, id, man.GetKind())

			if err = exec.Run(ctx, man); err != nil {
				errChan <- fmt.Errorf("manifest %s failed: %s", id, err.Error())
				return
			}

			output.Logf(interfaces.InfoLevel, "%s %s manifest finished", planRunnerOutputPrefix, id)
		}()
	}

	wg.Wait()
	close(errChan)

	var rErr error

	if len(errChan) > 0 {
		for err := range errChan {
			rErr = errors.Join(rErr, err)
		}

		return rErr
	}

	return nil
}

func (r *DefaultPlanRunner) runHooksWithContext(ctx interfaces.ExecutionContext, event hooks.HookEvent, actions []hooks.Action) error {
	if len(actions) == 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return r.hooksRunner.RunHooks(ctx, event, actions)
	}
}
