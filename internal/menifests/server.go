package menifests

type Server struct {
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
	Spec     struct {
		BaseUrl string            `yaml:"baseUrl"`
		Headers map[string]string `yaml:"headers"`
	} `yaml:"spec"`
}
