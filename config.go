package config

import (
	"errors"
	"os"
)

// Decoder is the interface that wraps the Decode method. Can be used to implement custom decoders.
type Decoder interface {
	Decode(val string) error
}

// Parser declare behavior to extend the different parsers that
// can be used to unmarshal conf.
type Parser interface {
	Parse(cfg interface{}) error
}

// source is the interface that wraps the Source method which is used to load the configuration
// from environment variables and command line flags.
// Source method accepts Field struct
type source interface {
	Source(f Field) (string, bool)
}

// MutatorFunc is a function that mutates a value of the key before it is set to the field.
type MutatorFunc func(key, value string) (string, error)

// Process processes the struct with environment variables and command line flags source.
func Process(prefix string, cfg interface{}, mutator ...MutatorFunc) error {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	return parseWithDefaultSource(prefix, args, cfg, mutator...)
}

// ProcessWithParser processes the struct with the given parsers. After processing with the parsers
// it will process the struct with environment variables and command line flags source.
// It also accepts mutator function to mutate the value before it is set to the field.
func ProcessWithParser(prefix string, cfg interface{}, parsers []Parser, mutator ...MutatorFunc) error {
	// command line arguments
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	// process the struct with the given parsers
	if err := processWithParser(cfg, parsers...); err != nil {
		return err
	}

	return parseWithDefaultSource(prefix, args, cfg, mutator...)
}

// processWithParser processes the struct with the given parsers.
func processWithParser(cfg interface{}, parsers ...Parser) error {
	for _, p := range parsers {
		if err := p.Parse(cfg); err != nil {
			return err
		}
	}

	return nil
}

// processWithSource processes the Field with the given source and mutator.
func processWithSource(f Field, source []source, mutator ...MutatorFunc) error {
	// iterate over the sources
	for _, src := range source {
		if src == nil {
			continue
		}

		// get the value from the source
		val, ok := src.Source(f)
		if !ok {
			continue
		}

		// if mutator is provided then execute the mutator
		// before setting the value to the field
		if len(mutator) > 0 {
			for _, m := range mutator {
				if m == nil {
					continue
				}

				var err error
				val, err = m(f.Name, val)
				if err != nil {
					return err
				}
			}
		}

		if err := processField(val, f.FieldValue); err != nil {
			return errors.New("error processing field: " + f.Name + ", error: " + err.Error())
		}

	}

	return nil
}

// parseWithDefaultSource parses the struct with environment variables and command line flags source.
// It also accepts mutator function to mutate the value before it is set to the field.
func parseWithDefaultSource(prefix string, args []string, cfg interface{}, mutator ...MutatorFunc) error {
	// create instance of flag source
	flag, err := newFlagParser(args)
	if err != nil {
		return err
	}

	sources := []source{newEnvSource(prefix), flag}

	// extract the fields from the given struct
	fields, err := extractFields(nil, cfg)
	if err != nil {
		return err
	}

	// iterate over the fields
	for _, f := range fields {
		// set the default value to the field if any
		// and make sure not to override the value if already set by Parser
		if f.Default != "" && f.FieldValue.IsZero() {
			if err := processField(f.Default, f.FieldValue); err != nil {
				return errors.New("error processing default value: " + f.Name + ", error: " + err.Error())
			}
		}

		// process the field with the given sources
		if err := processWithSource(f, sources, mutator...); err != nil {
			return err
		}

		// after processing the field at this point all the fields should be set
		// and if required field is not set then return error
		if f.Required && f.FieldValue.IsZero() {
			return errors.New("required field not set: " + f.Name)
		}
	}

	return nil

}
