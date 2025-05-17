package kinds

import (
	"fmt"
)

func FormManifestID(namespace, kind, name string) string {
	return fmt.Sprintf("%s.%s.%s", namespace, kind, name)
}
