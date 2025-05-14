package executor

import (
	"fmt"
	"github.com/apiqube/cli/core/plan"
	"github.com/apiqube/cli/plugins"
)

func ExecutePlan(plan *plan.ExecutionPlan) {
	for _, step := range plan.Steps {
		plugin, err := plugins.GetPlugin(step.Type)
		if err != nil {
			fmt.Printf("❌ Unknown plugin for step '%s'\n", step.Name)
			continue
		}
		fmt.Printf("🔧 Executing step: %s\n", step.Name)
		res, err := plugin.Execute(step, nil)
		if err != nil || !res.Success {
			fmt.Printf("❌ Step '%s' failed: %v\n", step.Name, err)
			continue
		}
		fmt.Printf("✅ Step '%s' passed\n", step.Name)
	}
}
