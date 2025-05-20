package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/apiqube/cli/internal/core/manifests"
)

func FormManifestID(namespace, kind, name string) string {
	return fmt.Sprintf("%s.%s.%s", namespace, kind, name)
}

func ParseManifestID(id string) (namespace, kind, name string) {
	parts := strings.Split(id, ".")

	switch len(parts) {
	case 3:
		namespace = parts[0]
		kind = parts[1]
		name = parts[2]
	case 2:
		namespace = manifests.DefaultNamespace
		kind = parts[0]
		name = parts[1]
	case 1:
		namespace = manifests.DefaultNamespace
		kind = parts[0]
		name = generateDefaultName()
	default:
		return "", "", ""
	}

	if kind == "" {
		return "", "", ""
	}

	return namespace, kind, name
}

func ParseManifestIDWithError(id string) (string, string, string, error) {
	namespace, kind, name := ParseManifestID(id)

	if kind == "" {
		return "", "", "", fmt.Errorf("manifest kind not specified: %s", id)
	}

	if name == "" {
		name = generateDefaultName()
	}

	if namespace == "" {
		namespace = manifests.DefaultNamespace
	}

	return namespace, kind, name, nil
}

func generateDefaultName() string {
	uuidStr := uuid.NewString()
	return "m-" + strings.Split(uuidStr, "-")[0]
}
