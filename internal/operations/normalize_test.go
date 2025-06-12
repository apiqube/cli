package operations

import (
	"sync"
	"testing"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
	"github.com/apiqube/cli/internal/core/manifests/kinds/servers"
	"github.com/apiqube/cli/internal/core/manifests/kinds/services"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/load"
	"github.com/apiqube/cli/internal/core/manifests/kinds/values"
	"github.com/apiqube/cli/internal/core/manifests/utils"
)

func hash(data []byte) string {
	h, _ := utils.CalculateContentHash(data)
	return h
}

func TestNormalize_StableHashes(t *testing.T) {
	testCases := []struct {
		name     string
		manifest manifests.Manifest
	}{
		{
			name: "Plan (realistic)",
			manifest: &plan.Plan{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.PlanKind,
					Metadata: kinds.Metadata{
						Name: "plan",
					},
				},
			},
		},
		{
			name: "Values",
			manifest: &values.Values{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.ValuesKind,
					Metadata: kinds.Metadata{
						Name: "values",
					},
				},
			},
		},
		{
			name: "Server",
			manifest: &servers.Server{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.ServerKind,
					Metadata: kinds.Metadata{
						Name: "server",
					},
				},
			},
		},
		{
			name: "Service",
			manifest: &services.Service{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.ServiceKind,
					Metadata: kinds.Metadata{
						Name: "service",
					},
				},
			},
		},
		{
			name: "HttpTest",
			manifest: &api.Http{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.HttpTestKind,
					Metadata: kinds.Metadata{
						Name: "http-test",
					},
				},
			},
		},
		{
			name: "HttpLoadTest",
			manifest: &load.Http{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.HttpLoadTestKind,
					Metadata: kinds.Metadata{
						Name: "http-load-test",
					},
				},
			},
		},
	}

	const runs = 3

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hashesYAML := make([]string, 0, runs)
			hashesJSON := make([]string, 0, runs)
			var wg sync.WaitGroup
			var mu sync.Mutex

			for i := 0; i < runs; i++ {
				wg.Add(2)
				go func() {
					defer wg.Done()
					data, err := NormalizeYAML(tc.manifest)
					if err != nil {
						t.Errorf("NormalizeYAML failed: %v", err)
						return
					}
					mu.Lock()
					hashesYAML = append(hashesYAML, hash(data))
					mu.Unlock()
				}()
				go func() {
					defer wg.Done()
					data, err := NormalizeJSON(tc.manifest)
					if err != nil {
						t.Errorf("NormalizeJSON failed: %v", err)
						return
					}
					mu.Lock()
					hashesJSON = append(hashesJSON, hash(data))
					mu.Unlock()
				}()
			}
			wg.Wait()
			for i := 1; i < len(hashesYAML); i++ {
				if hashesYAML[i] != hashesYAML[0] {
					t.Errorf("YAML hashes not equal for manifest %s: %v", tc.name, hashesYAML)
				}
			}

			for i := 1; i < len(hashesJSON); i++ {
				if hashesJSON[i] != hashesJSON[0] {
					t.Errorf("JSON hashes not equal for manifest %s: %v", tc.name, hashesJSON)
				}
			}
		})
	}
}
