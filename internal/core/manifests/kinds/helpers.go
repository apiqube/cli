package kinds

import (
	"encoding/json"
	"fmt"

	"github.com/apiqube/cli/internal/core/manifests"
	"gopkg.in/yaml.v3"
)

func FormManifestID(namespace, kind, name string) string {
	return fmt.Sprintf("%s.%s.%s", namespace, kind, name)
}

func BaseMarshalYAML(m manifests.Defaultable) ([]byte, error) {
	m.Default()
	return yaml.Marshal(m)
}

func BaseMarshalJSON(m manifests.Defaultable) ([]byte, error) {
	m.Default()
	return json.MarshalIndent(m, "", "  ")
}

func BaseUnmarshalYAML(bytes []byte, m manifests.Defaultable) error {
	m.Default()
	return yaml.Unmarshal(bytes, m)
}

func BaseUnmarshalJSON(bytes []byte, m manifests.Defaultable) error {
	m.Default()
	return json.Unmarshal(bytes, m)
}
