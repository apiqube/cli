package validate

import "github.com/apiqube/cli/internal/core/manifests"

type ValidatorErrorBuilder interface {
	AddError(detail ValidationErrorDetail)
	Build() error
}

type Validator interface {
	Validate(manifest manifests.Manifest) error
}

type ManifestsValidator interface {
	Validate(manifest ...manifests.Manifest) bool

	Valid() []manifests.Manifest
	Invalid() []manifests.Manifest
}
