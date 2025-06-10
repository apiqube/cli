package operations

import (
	"testing"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
	"github.com/goccy/go-yaml"

	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
	"github.com/apiqube/cli/internal/core/manifests/kinds/servers"
	"github.com/apiqube/cli/internal/core/manifests/kinds/services"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/load"
	"github.com/apiqube/cli/internal/core/manifests/kinds/values"
)

func TestParse_Manifests(t *testing.T) {
	testCases := []struct {
		name         string
		manifest     manifests.Manifest
		expectedType string
		expectErr    bool
		customData   []byte // for error/empty cases
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
			expectedType: manifests.PlanKind,
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
			expectedType: manifests.ValuesKind,
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
			expectedType: manifests.ServerKind,
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
			expectedType: manifests.ServiceKind,
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
			expectedType: manifests.HttpTestKind,
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
			expectedType: manifests.HttpLoadTestKind,
		},
		{
			name:       "UnknownKind",
			expectErr:  true,
			customData: []byte("kind: UnknownKind\napiVersion: v1\nmetadata:\n  name: test\n"),
		},
		{
			name:       "EmptyData",
			expectErr:  true,
			customData: []byte(""),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			var err error
			if tc.manifest != nil {
				data, err = yaml.Marshal(tc.manifest)
				if err != nil {
					t.Fatalf("Failed to marshal manifest: %v", err)
				}
			} else {
				data = tc.customData
			}
			m, err := Parse(YAMLFormat, data)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if m.GetKind() != tc.expectedType {
				t.Fatalf("Expected kind %s, got %s", tc.expectedType, m.GetKind())
			}
		})
	}
}
