package config

import (
	"os"
	"strings"
)

// env implements source interface for environment variables.
type env struct {
	m map[string]string
}

// newEnvSource returns a new Parser that can be used to process the conf struct with environment variables.
func newEnvSource(prefix string) source {
	if prefix != "" {
		prefix = strings.ToUpper(prefix) + "_"
	}

	// iterate over os.Environ and store the environment variables in a map
	m := make(map[string]string)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}

		idx := strings.Index(e, "=")
		m[strings.ToUpper(strings.TrimPrefix(e[:idx], prefix))] = e[idx+1:]
	}

	return &env{m: m}
}

func (e *env) Source(f Field) (string, bool) {
	// check if the key exists in the map
	if val, ok := e.m[f.EnvVar]; ok {
		return val, true
	}

	return "", false
}
