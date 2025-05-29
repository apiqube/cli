package validate

import (
	"testing"
	"time"

	"github.com/apiqube/cli/internal/core/manifests/kinds/services"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/load"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
	"github.com/apiqube/cli/internal/core/manifests/kinds/servers"
	"github.com/apiqube/cli/internal/core/manifests/kinds/values"
	"github.com/stretchr/testify/require"
)

var (
	invalidVersionManifest = &values.Values{
		BaseManifest: kinds.BaseManifest{
			Version: "notV1",
		},
		Spec: struct {
			Data map[string]any `yaml:",inline" json:",inline" validate:"required,min=1,dive"`
		}{Data: map[string]any{}},
	}

	invalidKindManifest = &values.Values{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    "NotValues",
		},
		Spec: struct {
			Data map[string]any `yaml:",inline" json:",inline" validate:"required,min=1,dive"`
		}{Data: map[string]any{}},
	}

	invalidNameManifest = &values.Values{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.ValuesKind,
		},
		Spec: struct {
			Data map[string]any `yaml:",inline" json:",inline" validate:"required,min=1,dive"`
		}{Data: map[string]any{}},
	}

	invalidPlanManifest = &plan.Plan{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.PlanKind,
			Metadata: kinds.Metadata{
				Name: "invalid_plan",
			},
		},
		Spec: struct {
			Stages []plan.Stage `yaml:"stages" json:"stages" validate:"required,min=1,dive"`
			Hooks  *plan.Hooks  `yaml:"hooks,omitempty" json:"hooks,omitempty" validate:"omitempty"`
		}{
			Stages: []plan.Stage{},
			Hooks:  &plan.Hooks{},
		},
	}

	invalidServiceManifest = &services.Service{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.ServiceKind,
			Metadata: kinds.Metadata{
				Name: "invalid_service",
			},
		},
		Spec: struct {
			Containers []services.Container `yaml:"containers" validate:"required,min=1,max=25,dive"`
		}{
			Containers: []services.Container{
				{
					Name:       "",
					Dockerfile: "Dockerfile",
					Image:      "Image",
					Replicas:   100,
				},
			},
		},
	}

	invalidServerManifest = &servers.Server{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.ServerKind,
			Metadata: kinds.Metadata{
				Name: "invalid_server",
			},
		},
		Spec: struct {
			BaseURL string            `yaml:"baseUrl" json:"baseUrl" validate:"required,url"`
			Health  string            `yaml:"health" json:"health" validate:"omitempty,max=100"`
			Headers map[string]string `yaml:"headers,omitempty" json:"headers" validate:"omitempty,max=20"`
		}{
			BaseURL: "",
			Health:  "",
			Headers: map[string]string{},
		},
	}

	invalidHttpTestManifest = &api.Http{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.HttpTestKind,
			Metadata: kinds.Metadata{
				Name: "invalid_http_test",
			},
		},
		Spec: struct {
			Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
			Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
		}{
			Target: "",
			Cases:  []api.HttpCase{},
		},
	}

	invalidHttpTestCaseManifest = &api.Http{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.HttpTestKind,
			Metadata: kinds.Metadata{
				Name: "invalid_http_test_case",
			},
		},
		Spec: struct {
			Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
			Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
		}{
			Target: "target",
			Cases: []api.HttpCase{
				{
					HttpCase: tests.HttpCase{
						Name:   "",
						Method: "INVALID",
					},
				},
			},
		},
	}

	invalidHttpLoadTestManifest = &load.Http{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.HttpLoadTestKind,
			Metadata: kinds.Metadata{
				Name: "invalid_http_load_test",
			},
		},
		Spec: struct {
			Target string          `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
			Cases  []load.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
		}{
			Target: "",
			Cases:  []load.HttpCase{},
		},
	}

	invalidHttpLoadTestCaseManifest = &load.Http{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.HttpLoadTestKind,
			Metadata: kinds.Metadata{
				Name: "invalid_http_load_test_case",
			},
		},
		Spec: struct {
			Target string          `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
			Cases  []load.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
		}{
			Target: "target",
			Cases: []load.HttpCase{
				{
					HttpCase: tests.HttpCase{
						Name:   "",
						Method: "INVALID",
					},
					Type:      "INVALID",
					Repeats:   -1,
					Agents:    -1,
					RPS:       -1,
					SaveEntry: -1,
				},
			},
		},
	}
)

var (
	validPlanManifest = &plan.Plan{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.PlanKind,
			Metadata: kinds.Metadata{
				Name: "valid_plan",
			},
		},
		Spec: struct {
			Stages []plan.Stage `yaml:"stages" json:"stages" validate:"required,min=1,dive"`
			Hooks  *plan.Hooks  `yaml:"hooks,omitempty" json:"hooks,omitempty" validate:"omitempty"`
		}{
			Stages: []plan.Stage{
				{
					Name: "stage_1",
					Manifests: []string{
						"id_0",
						"id_1",
					},
				},
			},
		},
	}

	validServiceManifest = &services.Service{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.ServiceKind,
			Metadata: kinds.Metadata{
				Name: "valid_service",
			},
		},
		Spec: struct {
			Containers []services.Container `yaml:"containers" validate:"required,min=1,max=25,dive"`
		}{
			Containers: []services.Container{
				{
					Name:       "some_name",
					Dockerfile: "Dockerfile",
				},
			},
		},
	}

	validServerManifest = &servers.Server{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.ServerKind,
			Metadata: kinds.Metadata{
				Name: "valid_server",
			},
		},
		Spec: struct {
			BaseURL string            `yaml:"baseUrl" json:"baseUrl" validate:"required,url"`
			Health  string            `yaml:"health" json:"health" validate:"omitempty,max=100"`
			Headers map[string]string `yaml:"headers,omitempty" json:"headers" validate:"omitempty,max=20"`
		}{
			BaseURL: "http://127.0.0.1:8080",
			Health:  "",
		},
	}

	validHttpTestCaseManifest = &api.Http{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.HttpTestKind,
			Metadata: kinds.Metadata{
				Name: "valid_http_test",
			},
		},
		Spec: struct {
			Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
			Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
		}{
			Target: "target",
			Cases: []api.HttpCase{
				{
					HttpCase: tests.HttpCase{
						Name:     "some_name_0",
						Method:   "GET",
						Endpoint: "/users",
						Url:      "http://127.0.0.1:8080/users",
						Headers: map[string]string{
							"Authorization": "Bearer token",
						},
						Body: map[string]any{
							"foo": "bar",
						},
						Assert: []*tests.Assert{
							{
								Target:   "status",
								Equals:   200,
								Contains: "some_data",
								Exists:   true,
								Template: "{{Status}} OK",
							},
						},
						Save: &tests.Save{
							Json: map[string]string{"foo": "bar"},
							Headers: map[string]string{
								"Authorization": "Bearer token",
							},
							Status: true,
							Body:   true,
							All:    true,
							Group:  "custom_save_group",
						},
						Pass: []tests.Pass{
							{
								From: "headers",
								Map:  map[string]string{"Authorization": "Bearer token"},
							},
						},
						Timeout:  time.Second,
						Parallel: false,
						Details: []string{
							"id_0",
							"id_1",
						},
					},
				},
			},
		},
	}

	validHttpLoadTestCaseManifest = &load.Http{
		BaseManifest: kinds.BaseManifest{
			Version: manifests.V1,
			Kind:    manifests.HttpLoadTestKind,
			Metadata: kinds.Metadata{
				Name: "valid_http_load_test",
			},
		},
		Spec: struct {
			Target string          `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
			Cases  []load.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
		}{
			Target: "target",
			Cases: []load.HttpCase{
				{
					HttpCase: tests.HttpCase{
						Name:   "case_0",
						Method: "POST",
					},
					Type: "wave",
					Wave: &load.WaveConfig{
						Low:   100,
						High:  10_000,
						Delta: 500,
					},
					Repeats:   10,
					Agents:    10,
					RPS:       1_000,
					SaveEntry: 5,
				},
			},
		},
	}
)

func TestValidator_Validate(t *testing.T) {
	v := NewValidator()

	var err error

	// INVALID
	t.Run("InvalidVersionManifest: incorrect version", func(t *testing.T) {
		err = v.Validate(invalidVersionManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "version")
	})

	t.Run("InvalidKindManifest: incorrect kind", func(t *testing.T) {
		err = v.Validate(invalidKindManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "one of")
	})

	t.Run("InvalidNameManifest: empty name", func(t *testing.T) {
		err = v.Validate(invalidNameManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "name")
	})

	t.Run("InvalidSpecDataManifest: spec data less then min", func(t *testing.T) {
		err = v.Validate(invalidNameManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data")
	})

	t.Run("InvalidPlanManifest: empty stages", func(t *testing.T) {
		err = v.Validate(invalidPlanManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "stages")
	})

	t.Run("InvalidServiceManifest: empty name, both fields, replicas to large", func(t *testing.T) {
		err = v.Validate(invalidServiceManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "name")
		require.Contains(t, err.Error(), "dockerfile")
		require.Contains(t, err.Error(), "image")
		require.Contains(t, err.Error(), "replicas")
	})

	t.Run("InvalidServerManifest: empty base url", func(t *testing.T) {
		err = v.Validate(invalidServerManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "baseurl")
	})

	t.Run("InvalidHttpTestManifest: empty target, not cases", func(t *testing.T) {
		err = v.Validate(invalidHttpTestManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "target")
		require.Contains(t, err.Error(), "cases")
	})

	t.Run("InvalidHttpTestManifest: invalid cases: empty name, invalid HTTP method", func(t *testing.T) {
		err = v.Validate(invalidHttpTestCaseManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "name")
	})

	t.Run("InvalidHttpLoadTestManifest: empty target, not cases", func(t *testing.T) {
		err = v.Validate(invalidHttpLoadTestManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "target")
		require.Contains(t, err.Error(), "cases")
	})

	t.Run("InvalidHttpLoadTestCaseManifest: load case not valid", func(t *testing.T) {
		err = v.Validate(invalidHttpLoadTestCaseManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "name")
		require.Contains(t, err.Error(), "type")
		require.Contains(t, err.Error(), "repeats")
		require.Contains(t, err.Error(), "agents")
		require.Contains(t, err.Error(), "rps")
		require.Contains(t, err.Error(), "saveentry")
	})

	// VALID
	t.Run("ValidPlanManifest", func(t *testing.T) {
		err = v.Validate(validPlanManifest)
		require.NoError(t, err)
	})

	t.Run("ValidServiceManifest", func(t *testing.T) {
		err = v.Validate(validServiceManifest)
		require.NoError(t, err)
	})

	t.Run("ValidServerManifest", func(t *testing.T) {
		err = v.Validate(validServerManifest)
		require.NoError(t, err)
	})

	t.Run("ValidHttpTestManifest", func(t *testing.T) {
		err = v.Validate(validHttpTestCaseManifest)
		require.NoError(t, err)
	})

	t.Run("ValidHttpLoadTestManifest", func(t *testing.T) {
		err = v.Validate(validHttpLoadTestCaseManifest)
		require.NoError(t, err)
	})
}
