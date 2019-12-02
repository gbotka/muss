package config

// DockerComposeConfig represents a docker-compose config file.
type DockerComposeConfig map[string]interface{}

// NewDockerComposeConfig returns new DockerComposeConfig
func NewDockerComposeConfig() DockerComposeConfig {
	return DockerComposeConfig{
		"version":  "3.7", // latest
		"volumes":  map[string]interface{}{},
		"services": map[string]interface{}{},
	}
}
