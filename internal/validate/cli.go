package validate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/ui"
)

var _ ManifestsValidator = (*ManifestValidator)(nil)

type ManifestValidator struct {
	validator      Validator
	ui             ui.UI
	valid, invalid []manifests.Manifest
}

func NewManifestValidator(validator Validator, ui ui.UI) *ManifestValidator {
	return &ManifestValidator{
		validator: validator,
		ui:        ui,
	}
}

func (v *ManifestValidator) Validate(manifests ...manifests.Manifest) bool {
	hasErrors := false
	errorBuilder := &ManifestErrorBuilder{ui: v.ui}

	for i, man := range manifests {
		errorBuilder.StartManifest(i+1, man.GetID())

		if err := v.validator.Validate(man); err != nil {
			hasErrors = true
			var vErr *ValidationError
			if errors.As(err, &vErr) {
				errorBuilder.AddValidationErrors(vErr.Details)
			} else {
				errorBuilder.AddGenericError(err)
			}

			v.invalid = append(v.invalid, man)
		} else {
			v.valid = append(v.valid, man)
		}

		errorBuilder.FinishManifest()
	}

	return !hasErrors
}

func (v *ManifestValidator) Valid() []manifests.Manifest {
	return v.valid
}

func (v *ManifestValidator) Invalid() []manifests.Manifest {
	return v.invalid
}

type ManifestErrorBuilder struct {
	ui         ui.UI
	manifestID string
	position   int
	hasErrors  bool
}

func (b *ManifestErrorBuilder) StartManifest(position int, id string) {
	b.position = position
	b.manifestID = id
	b.hasErrors = false
}

func (b *ManifestErrorBuilder) AddValidationErrors(details []ValidationErrorDetail) {
	b.hasErrors = true

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Manifest #%d - %s has validation errors:", b.position, b.manifestID))

	for _, detail := range details {
		builder.WriteString(fmt.Sprintf("\n- %-10s %s", detail.Field+":", detail.Message))
	}

	b.ui.Log(ui.TypeError, builder.String())
}

func (b *ManifestErrorBuilder) AddGenericError(err error) {
	b.hasErrors = true
	b.ui.Logf(ui.TypeError, "Manifest #%d - %s Error: %s",
		b.position,
		b.manifestID,
		err.Error(),
	)
}

func (b *ManifestErrorBuilder) FinishManifest() {
	if !b.hasErrors {
		b.ui.Logf(ui.TypeSuccess, "Manifest #%d - %s %s",
			b.position,
			b.manifestID,
			"✓ Valid",
		)
	}
}
