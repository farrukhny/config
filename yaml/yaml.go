// Package yaml provides yaml support by implementing the Parser interface.
package yaml

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAML provides support for unmarshalling YAML into the applications
// config value. After the yaml is unmarshalled, the Parse function is
// executed to apply value to config struct fields.
type YAML struct {
	data []byte
}

// WithData accepts the yaml document as a slice of bytes.
func WithData(data []byte) YAML {
	return YAML{
		data: data,
	}
}

// Reader accepts a reader to read the yaml.
func Reader(r io.Reader) YAML {
	var b bytes.Buffer
	if _, err := b.ReadFrom(r); err != nil {
		return YAML{}
	}

	return YAML{
		data: b.Bytes(),
	}
}

// Parse performs the actual processing of the yaml. It unmarshal the yaml into the config struct.
func (y YAML) Parse(cfg interface{}) error {
	err := yaml.Unmarshal(y.data, cfg)
	if err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}
	return nil
}
