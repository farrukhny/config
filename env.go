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
func newEnvSource() source {
	// iterate over os.Environ and store the environment variables in a map
	m := make(map[string]string)
	for _, e := range os.Environ() {
		// split the environment variable by "=" sign
		pair := strings.SplitN(e, "=", 2)
		// if the length of the pair is not equal to 2 then skip it
		if len(pair) != 2 {
			continue
		}
		// store the environment variable in the map
		m[pair[0]] = pair[1]
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
