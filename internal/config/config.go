package config

import "github.com/apiqube/cli/internal/test"

type Config struct {
	Version string      `yaml:"version"`
	Tests   []test.Case `yaml:"tests"`
}
