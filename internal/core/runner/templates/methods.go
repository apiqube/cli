package templates

import (
	"fmt"
	"strconv"
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

func methodClone(value any, args ...string) (any, error) {
	if len(args) > 0 {
		amount, err := strconv.Atoi(args[0])
		if err != nil {
			return value, nil
		}

		values := make([]any, amount)
		for i := 0; i < amount; i++ {
			values[i] = value
		}

		return values, nil
	}
	return value, nil
}

func methodReplace(value any, args ...string) (any, error) {
	if len(args)%2 != 0 {
		return value, nil
	}

	return strings.ReplaceAll(fmt.Sprint(value), strings.Trim(args[0], "'"), strings.Trim(args[1], "'")), nil
}
