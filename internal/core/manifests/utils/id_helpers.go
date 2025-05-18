package utils

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests"
	"strings"
)

func ParseManifestID(id string) (string, string, string) {
	parts := strings.Split(id, ".")
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2]
	} else if len(parts) == 2 {
		return manifests.DefaultNamespace, parts[0], parts[1]
	} else {
		return "", "", ""
	}
}

func ParseManifestIDWithError(id string) (string, string, string, error) {
	namespace, kind, name := ParseManifestID(id)
	if namespace == "" {
		namespace = manifests.DefaultNamespace
	} else if kind == "" {
		return namespace, name, name, fmt.Errorf("manifest kind not specified")
	} else if name == "" {
		return namespace, kind, name, fmt.Errorf("manifest name not specified")
	}
	return namespace, kind, name, nil
}
