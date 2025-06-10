package templates

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"strings"
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
