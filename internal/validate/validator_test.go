package validate

import (
	"testing"

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
			Headers: map[string]string{},
		},
	}
)

func TestValidator_Validate(t *testing.T) {
	v := NewValidator()

	var err error

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

	t.Run("InvalidServerManifest: empty base url", func(t *testing.T) {
		err = v.Validate(invalidServerManifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "baseurl")
	})

	t.Run("ValidServerManifest", func(t *testing.T) {
		err = v.Validate(validServerManifest)
		require.NoError(t, err)
	})
}
