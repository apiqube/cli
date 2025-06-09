package templates

import (
	"fmt"
	"strings"
)

func methodToUpper(value any, _ ...string) (any, error) {
	return strings.ToUpper(fmt.Sprintf("%v", value)), nil
}

func methodToLower(value any, _ ...string) (any, error) {
	return strings.ToLower(fmt.Sprintf("%v", value)), nil
}

func methodTrim(value any, args ...string) (any, error) {
	cutset := " \t\n\r"
	if len(args) > 0 {
		cutset = args[0]
	}

	return strings.Trim(fmt.Sprintf("%v", value), cutset), nil
}

func methodToString(value any, _ ...string) (any, error) {
	return fmt.Sprintf("%v", value), nil
}
