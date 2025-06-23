package kinds

import "github.com/apiqube/cli/internal/core/manifests"

var (
	PriorityMap = map[string]int{
		// Infrastructure kinds
		manifests.ValuesKind: 10,

		// Application kinds
		manifests.ServerKind:  100,
		manifests.ServiceKind: 110,

		// Test kinds
		manifests.HttpTestKind: 200,

		// Load test kinds
		manifests.HttpLoadTestKind: 300,
	}
)
