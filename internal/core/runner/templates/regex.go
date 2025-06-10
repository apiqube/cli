package templates

import (
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

func regex(args ...string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("please provide a valid regex pattern")
	}

	pattern := args[0]
	pattern = strings.Trim(pattern, `"'`)
	pattern = strings.ReplaceAll(pattern, `\\`, `\`)
	pattern = strings.ReplaceAll(pattern, `\"`, `"`)
	pattern = strings.ReplaceAll(pattern, `\'`, `'`)

	return gofakeit.Regex(pattern), nil
}
