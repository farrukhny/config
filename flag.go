package config

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrHelp is returned when the help flag is set.
	ErrHelp = errors.New("help requested")
	// ErrVersion is returned when the version flag is set.
	ErrVersion = errors.New("version requested")
)

type flagValue struct {
	HasValue bool
	Value    string
}

// flag implements the Parser interface for command line arguments.
type flag struct {
	args map[string]flagValue
}

// newFlagParser returns a new Parser that can be used to process the conf struct with command line arguments.
func newFlagParser(args []string) (source, error) {
	m := make(map[string]flagValue)
	if len(args) > 0 {
		for i := 0; i < len(args); i++ {
			s := args[i]

			// if argument too short or doesn't start with a dash "-" then break
			if len(s) < 2 || s[0] != '-' {
				break
			}

			minus := 1
			// if the argument starts with two dashes "--" then increment minus by 1
			if s[1] == '-' {
				minus++
			}

			// assign the flag name
			name := s[minus:]
			// check if name is not empty or starts with a dash "-" or starts with equal "="
			if len(name) == 0 || name[0] == '-' || name[0] == '=' {
				return nil, fmt.Errorf("bad flag syntax: %s", s)
			}

			// if flag has a value after "=" sign then use it as value for the flag
			// otherwise use the next argument as value
			// for example: --flag=value or --flag value
			hasValue := false
			value := ""
			for i := 0; i < len(name); i++ {
				if name[i] == '=' {
					value = name[i+1:]
					name = name[:i]
					hasValue = true
					break
				}
			}

			if name == "help" || name == "h" {
				return nil, ErrHelp
			}

			if name == "version" || name == "v" {
				return nil, ErrVersion
			}

			// if the flag still not have a value then use the next argument as value
			// flag maybe in form of --flag value
			if !hasValue && i+1 < len(args) {
				if args[i+1][0] != '-' {
					hasValue = true
					value = args[i+1]
					i++
				}
			}

			// store the key and value in the map
			m[name] = flagValue{
				HasValue: hasValue,
				Value:    value,
			}
		}
	}

	return &flag{args: m}, nil
}

// Source will return the value of the key if found.
func (f *flag) Source(field Field) (string, bool) {
	var isBoolType = field.FieldValue.Kind() == reflect.Bool

	if field.ShortFlag != 0 {
		if val, ok := f.source(string(field.ShortFlag), isBoolType); ok {
			return val, true
		}
	}

	return f.source(field.Flag, isBoolType)
}

func (f *flag) source(key string, isBool bool) (string, bool) {
	val, ok := f.args[key]
	if !ok || !isBool {
		return val.Value, ok
	}

	if val.HasValue {
		return val.Value, ok
	}

	if !val.HasValue && isBool {
		return "true", ok
	}

	return "", false
}
