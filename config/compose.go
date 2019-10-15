package config

import (
	"fmt"

	homedir "github.com/mitchellh/go-homedir"
)

// DockerComposeFile is the path to the docker-compose file that will be
// generated.
var DockerComposeFile = "docker-compose.yml"

// DockerComposeConfig returns an object ready to be yaml-dumped as a
// docker-compose file (or an error).
func DockerComposeConfig(config ProjectConfig) (map[string]interface{}, error) {
	dcc, _, err := DockerComposeFiles(config)
	return dcc, err
}

// DockerComposeFiles returns a map of the docker-compose config,
// a map representing supplementary files, and an error.
// The files value is a `map[string]func(string) error` where the key is
// the file path and the value is a function that takes the path argument
// and writes the file or errors.
func DockerComposeFiles(config ProjectConfig) (map[string]interface{}, map[string]func(string) error, error) {
	files := make(map[string]func(string) error)

	var user map[string]interface{}
	if u, ok := config["user"].(map[string]interface{}); ok {
		user = u
	}

	// Setup a base to merge things onto.
	dcc := map[string]interface{}{
		"version":  "3.7", // latest
		"volumes":  map[string]interface{}{},
		"services": map[string]interface{}{},
	}

	var servdefs []ServiceDef
	if val, ok := config["service_definitions"].([]ServiceDef); ok {
		servdefs = val
	} else if val, ok := config["service_definitions"].([]interface{}); ok {
		servdefs = make([]ServiceDef, len(val))
		for i, s := range val {
			servdefs[i] = s.(map[string]interface{})
		}
	}

	for _, service := range servdefs {
		servconf, err := serviceConfig(config, service)
		if err != nil {
			return nil, nil, err
		}

		dcc = mapMerge(dcc, servconf)
	}

	if override, ok := user["override"].(map[string]interface{}); ok {
		dcc = mapMerge(dcc, override)
	}

	if services, ok := (dcc["services"]).(map[string]interface{}); ok {
		for name, si := range services {
			if service, ok := si.(map[string]interface{}); ok {

				filevols, err := findFileVolumes(service)
				if err != nil {
					return nil, nil, err
				}
				for _, filevol := range filevols {
					files[filevol] = ensureFile
				}

				if !isValidService(service) {
					delete(services, name)
				}

			}
		}
	}

	return dcc, files, nil
}

func serviceConfig(config map[string]interface{}, service ServiceDef) (map[string]interface{}, error) {
	serviceName := service["name"].(string)
	serviceConfigs := service["configs"].(map[string]interface{})
	options := mapKeys(serviceConfigs)

	var userConfig map[string]interface{}
	if user, ok := config["user"].(map[string]interface{}); ok {
		userConfig = user
	}

	var result map[string]interface{}

	// Check if user configured a specific choice for this service:
	// `services: {somename: {config: choice}}`
	userChoice := ""
	if userServices, ok := userConfig["services"].(map[string]interface{}); ok {
		if userserv, ok := userServices[serviceName].(map[string]interface{}); ok {
			if val, ok := userserv["config"].(string); ok {
				userChoice = val
				if _, ok := serviceConfigs[userChoice]; !ok {
					return nil, fmt.Errorf("Config '%s' for service '%s' does not exist", userChoice, serviceName)
				}
			}
		}
	}

	if userChoice != "" {
		// If user chose specifically, use it.
		result = serviceConfigs[userChoice].(map[string]interface{})
	} else if len(options) == 1 {
		// If there is only one option, use it.
		result = serviceConfigs[options[0]].(map[string]interface{})
	} else {

		// To determine which config option to use we can build a list...
		var order []string

		// starting with any user configured preference
		if slice, ok := stringSlice(userConfig["service_preference"]); ok {
			order = append(order, slice...)
		}

		// followed by any project defaults
		if slice, ok := stringSlice(config["default_service_preference"]); ok {
			order = append(order, slice...)
		}

		// then iterate and use the first preference that this service defines.
		for _, o := range order {
			if found, ok := serviceConfigs[o]; ok {
				result = found.(map[string]interface{})
				break
			}
		}

	}
	// TODO: recurse
	if includes, ok := result["include"].([]interface{}); ok {
		delete(result, "include")
		base := map[string]interface{}{}
		for _, i := range includes {
			base = mapMerge(base, serviceConfigs[i.(string)].(map[string]interface{}))
		}
		result = mapMerge(base, result)
	}
	return result, nil
}

func isValidService(service map[string]interface{}) bool {
	if _, ok := service["build"]; ok {
		return true
	}
	if _, ok := service["image"]; ok {
		return true
	}
	return false
}

// Look for volumes marked as "file" (muss extension) and make sure they are
// files to avoid the terrible confusion that ensues when docker creates a
// directory where you expected a file.
// > NOTE: File must exist, else "It is always created as a directory".
// > https://docs.docker.com/storage/bind-mounts/#differences-between--v-and---mount-behavior
func findFileVolumes(service map[string]interface{}) ([]string, error) {
	var filevols []string
	if volumes, ok := service["volumes"].([]interface{}); ok {
		for _, volume := range volumes {
			if v, ok := volume.(map[string]interface{}); ok {
				if v["type"] == "bind" {
					if file, ok := v["file"].(bool); ok && file {
						expanded, err := homedir.Expand(expand(v["source"].(string)))
						if err != nil {
							return nil, err
						}
						filevols = append(filevols, expanded)
						delete(v, "file")
					}
				}
			}
		}
	}
	return filevols, nil
}

var keysToOverwrite = []string{"entrypoint", "command"}

func mapMerge(target map[string]interface{}, source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(target)+len(source))
	for k, v := range target {
		result[k] = v
	}
	for k, v := range source {
		if !mapMergeOverwrites(k) {
			if current, ok := result[k]; ok {
				if m, ok := current.(map[string]interface{}); ok {
					result[k] = mapMerge(m, v.(map[string]interface{}))
					continue
				} else if s, ok := current.([]interface{}); ok {
					vs := v.([]interface{})
					tmp := make([]interface{}, 0, len(s)+len(vs))
					tmp = append(tmp, s...)
					tmp = append(tmp, vs...)
					result[k] = tmp
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}

func mapMergeOverwrites(k string) bool {
	for _, o := range keysToOverwrite {
		if o == k {
			return true
		}
	}
	return false
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k[0:1] != "_" {
			keys = append(keys, k)
		}
	}
	return keys
}
