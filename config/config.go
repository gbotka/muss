package config

// ProjectFile is the path to the project config file.
var ProjectFile = "muss.yaml"

// UserFile is the path to the user config file, if defined.
var UserFile string

var project *ProjectConfig

// All returns the whole project config.
func All() (*ProjectConfig, error) {
	if project == nil {
		err := load()
		if err != nil {
			return nil, err
		}
	}
	return project, nil
}
