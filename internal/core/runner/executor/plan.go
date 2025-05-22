package executor

import (
	"errors"
	"fmt"
	"sync"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
	"github.com/apiqube/cli/internal/core/runner/hooks"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

const planRunnerOutputPrefix = "Plan Runner:"

var _ interfaces.PlanRunner = (*DefaultPlanRunner)(nil)

type DefaultPlanRunner struct {
	ExecutorRegistry interfaces.ExecutorRegistry
	HooksRunner      hooks.Runner
}

func NewDefaultPlanRunner(registry interfaces.ExecutorRegistry, hooksRunner hooks.Runner) *DefaultPlanRunner {
	return &DefaultPlanRunner{
		ExecutorRegistry: registry,
		HooksRunner:      hooksRunner,
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

	if p.Spec.Hooks != nil {
		if err = r.HooksRunner.RunHooks(ctx, hooks.BeforeRun, p.Spec.Hooks.BeforeRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan before start hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}
	}

	for _, stage := range p.Spec.Stages {
		stageName := stage.Name
		output.Logf(interfaces.InfoLevel, "%s %s stage starting...", planRunnerOutputPrefix, stageName)

		if err = r.HooksRunner.RunHooks(ctx, hooks.BeforeRun, stage.Hooks.BeforeRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s stage %s before start hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
			return err
		}

		var execErr error
		if stage.Parallel {
			execErr = r.runManifestsParallel(ctx, stage.Manifests)
		} else {
			execErr = r.runManifestsStrict(ctx, stage.Manifests)
		}

		if err = r.HooksRunner.RunHooks(ctx, hooks.AfterRun, stage.Hooks.AfterRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s stage %s after finish hooks running failed: %s", planRunnerOutputPrefix, stageName, err.Error())
			return err
		}

		if execErr != nil {
			output.Logf(interfaces.ErrorLevel, "%s stage %s failed\nReason: %s", planRunnerOutputPrefix, stageName, execErr.Error())

			if err = r.HooksRunner.RunHooks(ctx, hooks.OnFailure, stage.Hooks.OnFailure); err != nil {
				output.Logf(interfaces.ErrorLevel, "%s stage %s on failure hooks running failed\nReason: %s", planRunnerOutputPrefix, stageName, err.Error())
				return err
			}

			if p.Spec.Hooks != nil {
				if err = r.HooksRunner.RunHooks(ctx, hooks.OnFailure, p.Spec.Hooks.OnFailure); err != nil {
					output.Logf(interfaces.ErrorLevel, "%s plan on failure hooks runnin failed\nReason: %s", planRunnerOutputPrefix, err.Error())
					return errors.Join(execErr, err)
				}
			}

			return execErr
		}

		if err = r.HooksRunner.RunHooks(ctx, hooks.OnSuccess, stage.Hooks.OnSuccess); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan on success hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}
	}

	if p.Spec.Hooks != nil {
		if err = r.HooksRunner.RunHooks(ctx, hooks.AfterRun, p.Spec.Hooks.AfterRun); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan after finish hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}

		if err = r.HooksRunner.RunHooks(ctx, hooks.OnSuccess, p.Spec.Hooks.OnSuccess); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s plan on success hooks running failed\nReason: %s", planRunnerOutputPrefix, err.Error())
			return err
		}
	}

	output.Logf(interfaces.InfoLevel, "%s plan finished", planID)

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

		exec, exists := r.ExecutorRegistry.Find(man.GetKind())
		if !exists {
			return fmt.Errorf("no executor found for kind: %s", man.GetKind())
		}

		output.Logf(interfaces.InfoLevel, "%s %s running manifest using executor for: %s", planRunnerOutputPrefix, id, man.GetKind())

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

			exec, exists := r.ExecutorRegistry.Find(man.GetKind())
			if !exists {
				errChan <- fmt.Errorf("no executor found for kind: %s", man.GetKind())
				return
			}

			output.Logf(interfaces.InfoLevel, "%s %s running manifest using executor for: %s", planRunnerOutputPrefix, id, man.GetKind())

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
