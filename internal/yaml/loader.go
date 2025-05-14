package yaml

import (
	"gopkg.in/yaml.v3"
	"os"
)

func LoadConfig[T any](filePath string) (*T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var t T
	if err = yaml.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}
