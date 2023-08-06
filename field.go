package config

import (
	"encoding"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrInvalidTarget = errors.New("targetStruct must be a non-nil pointer")
)

const (
	envVarTag        = "env"
	defaultValueTag  = "default"
	requiredValueTag = "required"
	flagTag          = "flag"
	shortFlagTag     = "shortFlag"
	usageTag         = "usage"
	maskTag          = "mask"
	delimiter        = ","
	separator        = ":"
)

type Field struct {
	FieldValue reflect.Value
	Name       string
	EnvVar     string
	Flag       string
	ShortFlag  rune
	Default    string
	Required   bool
	Mask       bool
	Usage      string
}

// extractFields parses the struct and returns the list of Fields.
func extractFields(prefix []string, targetStruct interface{}) ([]Field, error) {
	if prefix == nil {
		prefix = []string{}
	}

	fields := make([]Field, 0)

	v := reflect.ValueOf(targetStruct)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, ErrInvalidTarget
	}

	// Dereference the pointer
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return nil, ErrInvalidTarget
	}

	t := e.Type()
	for i := 0; i < t.NumField(); i++ {
		f := e.Field(i)
		sf := t.Field(i)

		// Ignore unexported fields
		if !f.CanSet() || sf.PkgPath != "" || sf.Tag.Get(envVarTag) == "-" {
			continue
		}

		envVar := sf.Tag.Get(envVarTag)
		flagName := sf.Tag.Get(flagTag)
		shortFlag := sf.Tag.Get(shortFlagTag)
		defaultValue := sf.Tag.Get(defaultValueTag)
		requiredValue := sf.Tag.Get(requiredValueTag)
		maskValue := sf.Tag.Get(maskTag)
		usageValue := sf.Tag.Get(usageTag)

		fieldName := sf.Name
		fieldKey := append(prefix, splitCamelCase(fieldName)...)

		envName, err := createOrValidateEnvVarName(envVar, fieldKey)
		if err != nil {
			return nil, err
		}

		flag, err := createOrValidateFlagName(flagName, fieldKey)
		if err != nil {
			return nil, err
		}

		// Validate short flag name
		var short rune
		if len(shortFlag) > 0 {
			if len([]rune(shortFlag)) != 1 {
				return nil, errors.New("short flag name must be a single character")
			}
			short = []rune(shortFlag)[0]
		}

		// Check if field is required and has a default value
		if requiredValue == "true" && defaultValue != "" {
			return nil, fmt.Errorf("required field %s cannot have a default value", fieldName)
		}

		field := Field{
			FieldValue: f,
			Name:       strings.Join(fieldKey, "_"),
			EnvVar:     envName,
			Flag:       flag,
			ShortFlag:  short,
			Default:    defaultValue,
			Required:   requiredValue == "true",
			Mask:       maskValue == "true",
			Usage:      usageValue,
		}

		fields = append(fields, field)

		// Drill down through struct fields
		if f.Kind() == reflect.Struct {
			// Process value with decoder
			if err := processDecoder("", f); err != nil {
				return nil, errors.New("error processing decoder for FieldValue: " + sf.Name + " " + err.Error())
			}

			innerPrefix := fieldKey
			if sf.Anonymous {
				innerPrefix = prefix
			}

			embeddedPtr := f.Addr().Interface()
			embeddedFields, err := extractFields(innerPrefix, embeddedPtr)
			if err != nil {
				return nil, errors.New("error parsing embedded struct for FieldValue: " + sf.Name + " " + err.Error())
			}
			fields = append(fields[:len(fields)-1], embeddedFields...)
			continue
		}

	}

	return fields, nil
}

// processDecoder processes the Field as a decoder or custom unmarshaler.
func processDecoder(value string, field reflect.Value) error {
	if field.CanAddr() {
		field = field.Addr()
	}

	iface := field.Interface()
	switch iface := iface.(type) {
	case Decoder:
		return iface.Decode(value)
	case encoding.TextUnmarshaler:
		return iface.UnmarshalText([]byte(value))
	case json.Unmarshaler:
		return iface.UnmarshalJSON([]byte(value))
	case encoding.BinaryUnmarshaler:
		return iface.UnmarshalBinary([]byte(value))
	case gob.GobDecoder:
		return iface.GobDecode([]byte(value))
	}

	return nil
}

// processField processes the Field as a string.
func processField(value string, field reflect.Value) error {
	// handle pointers and uninitialized pointers
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("error parsing bool: " + err.Error())
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			val int64
			err error
		)
		// special case for time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, field.Type().Bits())
		}
		if err != nil {
			return errors.New("error parsing int: " + err.Error())
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 0, field.Type().Bits())
		if err != nil {
			return errors.New("error parsing uint: " + err.Error())
		}
		field.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return errors.New("error parsing float: " + err.Error())
		}
		field.SetFloat(i)
	case reflect.Slice:
		// special case for []byte
		if field.Type().Elem().Kind() == reflect.Uint8 {
			field.Set(reflect.ValueOf([]byte(value)))
		} else {
			vals := strings.Split(value, delimiter)
			s := reflect.MakeSlice(field.Type(), len(vals), len(vals))
			for i, v := range vals {
				v = strings.TrimSpace(v)
				if err := processField(v, s.Index(i)); err != nil {
					return err
				}
			}
			field.Set(s)
		}
	case reflect.Map:
		vals := strings.Split(value, delimiter)
		mp := reflect.MakeMapWithSize(field.Type(), len(vals))
		for _, v := range vals {
			kv := strings.SplitN(v, separator, 2)
			if len(kv) != 2 {
				return errors.New("invalid map value: " + v)
			}

			mKey, mVal := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])

			k := reflect.New(field.Type().Key()).Elem()
			if err := processField(mKey, k); err != nil {
				return err
			}

			v := reflect.New(field.Type().Elem()).Elem()
			if err := processField(mVal, v); err != nil {
				return err
			}

			mp.SetMapIndex(k, v)
		}
		field.Set(mp)
	default:
		return errors.New("unsupported type " + field.Type().String())
	}

	return nil
}

// valueToString accepts a reflect.Value and returns a string representation of it.
func valueToString(v reflect.Value) string {
	if v.IsValid() {
		switch v.Kind() {
		case reflect.String:
			return v.String()
		case reflect.Bool:
			return strconv.FormatBool(v.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// special case for time.Duration
			if v.Type() == reflect.TypeOf(time.Duration(0)) {
				return v.Interface().(time.Duration).String()
			}
			return strconv.FormatInt(v.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(v.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits())
		case reflect.Slice:
			// special case for []byte
			if v.Type().Elem().Kind() == reflect.Uint8 {
				return string(v.Bytes())
			}
			var vals []string
			for i := 0; i < v.Len(); i++ {
				vals = append(vals, valueToString(v.Index(i)))
			}
			return strings.Join(vals, delimiter)
		case reflect.Map:
			var vals []string
			for _, k := range v.MapKeys() {
				vals = append(vals, valueToString(k)+separator+valueToString(v.MapIndex(k)))
			}
			return strings.Join(vals, delimiter)
		default:
			return ""
		}
	}

	return ""
}

// createOrValidateEnvVarName validate env var that been given with a tag, if it is empty will generate default env var name from filed name.
// It will return error if env var name is invalid.
func createOrValidateEnvVarName(envVarTag string, filedKey []string) (string, error) {
	if envVarTag == "" {
		return strings.ToUpper(strings.Join(filedKey, "_")), nil
	}

	if !validateEnvVarName(envVarTag) {
		return "", errors.New("invalid environment variable name has been provided: " + envVarTag)
	}

	return envVarTag, nil
}

// createOrValidateFlagName validate flag that been given with a tag, if it is empty will generate default flag name from filed name.
// It will return error if flag name is invalid.
func createOrValidateFlagName(flagTag string, filedKey []string) (string, error) {
	if flagTag == "" {
		return strings.ToLower(strings.Join(filedKey, "-")), nil
	}

	if !validateFlagName(flagTag) {
		return "", errors.New("invalid flag name has been provided: " + flagTag)
	}

	return flagTag, nil
}

// validateEnvVarName validates given string as a valid environment variable name.
// Per IEEE Std 1003.1-2001, the name of an environment variable shall consist
// only of uppercase letters, digits, and the '_' (underscore) from the characters.
// The first character of an environment variable name shall not be a digit.
func validateEnvVarName(name string) bool {
	if len(name) == 0 {
		return false
	}

	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	for _, ch := range name {
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			continue
		}
		return false
	}

	return true
}

// validateFlagName validates given string as a valid flag name.
// flag name should be all in lower case. It can contain only letters, numbers, and hyphens.
// It should start with a letter.
// It should not end with a hyphen. It should not contain two consecutive hyphens.
// for example, "foo-bar", "foo-bar-1", "foo1-bar" are valid flag names.
func validateFlagName(name string) bool {
	if len(name) == 0 {
		return false
	}

	if name[0] < 'a' || name[0] > 'z' {
		return false
	}

	if name[len(name)-1] == '-' {
		return false
	}

	for i, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			if ch == '-' && name[i-1] == '-' {
				return false
			}
			continue
		}
		return false
	}

	return true
}

// TODO: some combination of words not working, like "APIKey" can not be parsed correctly.
// splitCamelCase splits camel case string and returns slice of words. It will use rune to split words.
// For example, "MyVar" -> []string{"My", "Var"}
func splitCamelCase(s string) []string {
	if s == "" {
		return []string{}
	}

	runes := []rune(s)
	lastChar := runes[0]
	lastIndex := 0
	var words []string

	for i, char := range runes {
		if unicode.IsUpper(char) && !unicode.IsUpper(lastChar) {
			words = append(words, string(runes[lastIndex:i]))
			lastIndex = i
		}
		lastChar = char
	}

	words = append(words, string(runes[lastIndex:]))

	return words
}
