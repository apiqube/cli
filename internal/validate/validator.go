package validate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/go-playground/validator/v10"
)

func NewValidator() Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	for kind, validFunc := range manifestKinsValidationFuncs {
		_ = v.RegisterValidation(kind, validFunc)
	}

	for name, validFunc := range validationFuncs {
		_ = v.RegisterValidation(name, validFunc)
	}

	return &baseValidator{
		validate: v,
	}
}

type baseValidator struct {
	validate *validator.Validate
}

func (b *baseValidator) Validate(manifest manifests.Manifest) error {
	if err := b.validate.Struct(manifest); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			errBuilder := &ValidationErrorBuilder{
				Errors: make([]ValidationErrorDetail, 0, len(validationErrors)),
			}

			for _, fieldErr := range validationErrors {
				errBuilder.AddError(ValidationErrorDetail{
					Field:   fieldErr.Field(),
					Tag:     fieldErr.Tag(),
					Value:   fieldErr.Value(),
					Message: buildErrorMessage(fieldErr),
				})
			}

			return errBuilder.Build()
		}

		return fmt.Errorf("manifest validation error: %w", err)
	}
	return nil
}

type ValidationErrorBuilder struct {
	Errors []ValidationErrorDetail
}

type ValidationError struct {
	Details []ValidationErrorDetail
}

type ValidationErrorDetail struct {
	Field   string
	Tag     string
	Value   any
	Message string
}

func (b *ValidationErrorBuilder) AddError(detail ValidationErrorDetail) {
	b.Errors = append(b.Errors, detail)
}

func (b *ValidationErrorBuilder) Build() error {
	if len(b.Errors) == 0 {
		return nil
	}
	return &ValidationError{Details: b.Errors}
}

func (e *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString("validation failed:\n")
	for i, detail := range e.Details {
		sb.WriteString(fmt.Sprintf("%d) %s\n", i+1, detail.Message))
	}
	return sb.String()
}

func buildErrorMessage(fieldErr validator.FieldError) string {
	fieldName := fieldErr.Field()

	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("field '%s' is required", fieldName)
	case "min":
		return fmt.Sprintf("field '%s' must be at least %s", fieldName, fieldErr.Param())
	case "max":
		return fmt.Sprintf("field '%s' must be at most %s", fieldName, fieldErr.Param())
	case "email":
		return fmt.Sprintf("field '%s' must be a valid email", fieldName)
	case "oneof":
		return fmt.Sprintf("field '%s' must contain only one of: %s", fieldName, strings.Join(strings.Split(strings.TrimSpace(fieldErr.Param()), " "), " | "))
	case "excluded_with":
		return fmt.Sprintf("field '%s' must be exclusive between: %s and %s", fieldName, fieldName, strings.Join(strings.Split(strings.TrimSpace(fieldErr.Param()), " "), " | "))
	case "eq":
		return fmt.Sprintf("field '%s' must be equal to %s", fieldName, fieldErr.Param())
	default:
		return fmt.Sprintf("field '%s' failed validation '%s'", fieldName, fieldErr.Tag())
	}
}
