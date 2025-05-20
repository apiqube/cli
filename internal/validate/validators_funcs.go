package validate

import (
	"strings"
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/go-playground/validator/v10"
)

var (
	validationFuncs = map[string]func(fl validator.FieldLevel) bool{
		"duration": func(fl validator.FieldLevel) bool {
			_, err := time.ParseDuration(fl.Field().String())
			return err == nil
		},
		"contains_template": func(fl validator.FieldLevel) bool {
			value := fl.Field().String()
			return strings.Contains(value, "{") && strings.Contains(value, "}")
		},
	}

	manifestKinsValidationFuncs = map[string]func(fl validator.FieldLevel) bool{
		manifests.PlanManifestKind: planValidationFunc,
	}
)

func planValidationFunc(fl validator.FieldLevel) bool {
	params, ok := fl.Field().Interface().(map[string]any)
	if !ok {
		return false
	}

	actionType := fl.Parent().FieldByName("Type").String()

	switch actionType {
	case "exec":
		if _, ok = params["command"]; !ok {
			return false
		}
	case "notify":
		if _, ok = params["target"]; !ok {
			return false
		}
	}
	return true
}
