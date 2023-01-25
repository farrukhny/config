package config

import (
	"encoding"
	"encoding/gob"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrInvalidTarget = errors.New("invalid target, must be a pointer to a struct")
)

// struct tags as constants
const (
	// envVarTag tag name for environment variable name
	envVarTag = "env"
	// default tag name for default value
	defaultValueTag = "default"
	// requiredValueTag tag name for requiredValueTag value
	requiredValueTag = "required"
	// flagTag tag name for flagTag name which will
	flagTag = "flag"
	// shortFlagTag tag name for short flag name which will
	shortFlagTag = "shortFlag"
	// usageTag tag name for description of the flagTag and envVarTag var
	usageTag = "usage"
	// maskTag tag name for masking the value of struct filed
	maskTag = "mask"

	delimiter = ","
	separator = ":"
)

// Field will be used to store the information about the Field
type Field struct {
	// FieldValue value
	FieldValue reflect.Value
	// FieldType is the type of the FieldValue
	FieldType reflect.Kind
	// Name of the FieldValue
	Name string
	// EnvVar of the FieldValue
	EnvVar string
	// flag name
	Flag string
	// short flag name
	ShortFlag rune
	// Default value of the FieldValue
	Default string
	// required value
	Required bool
	// mask value of the FieldValue in the output
	Mask bool
	// usage value
	Usage string
}

// extractFields parses the struct and returns the list of Fields.
func extractFields(prefix []string, targetStruct interface{}) ([]Field, error) {
	if prefix == nil {
		prefix = []string{}
	}

	var fields []Field

	s := reflect.ValueOf(targetStruct)
	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidTarget
	}

	// Dereference the pointer
	e := s.Elem()
	if e.Kind() != reflect.Struct {
		return nil, ErrInvalidTarget
	}

	targetType := e.Type()
	for i := 0; i < targetType.NumField(); i++ {
		f := e.Field(i)
		structField := targetType.Field(i)
		envName := structField.Tag.Get(envVarTag)
		flagName := structField.Tag.Get(flagTag)
		shortFlag := strings.ToLower(structField.Tag.Get(shortFlagTag))
		defaultValue := structField.Tag.Get(defaultValueTag)
		requiredValue := strings.ToLower(structField.Tag.Get(requiredValueTag))
		maskValue := strings.ToLower(structField.Tag.Get(maskTag))
		usageValue := structField.Tag.Get(usageTag)

		// Ignore Fields
		if !f.CanSet() || envName == "-" {
			continue
		}

		// if field is required and has default value then it returns error
		if requiredValue == "true" && defaultValue != "" {
			return nil, errors.New("required field can not have default value: " + structField.Name)
		}

		// drill down through pointers
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				//if is not struct skip it
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}

				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		// FieldValue key
		filedName := structField.Name
		fieldKey := append(prefix, splitCamelCase(filedName)...)

		env, err := createOrValidateEnvVarName(envName, fieldKey)
		if err != nil {
			return nil, err
		}
		envName = env

		flag, err := createOrValidateFlagName(flagName, fieldKey)
		if err != nil {
			return nil, err
		}
		flagName = flag

		// if short flag name is not single character, return error
		if shortFlag != "" && len([]rune(shortFlag)) != 1 {
			return nil, errors.New("short flag name must be single character")
		}
		short := []rune(shortFlag)
		if len(short) == 0 {
			short = []rune{0}
		}

		fields = append(fields, Field{
			FieldValue: f,
			FieldType:  f.Kind(),
			Name:       strings.Join(fieldKey, "_"),
			EnvVar:     envName,
			Flag:       flagName,
			ShortFlag:  short[0],
			Default:    defaultValue,
			Required:   requiredValue == "true",
			Mask:       maskValue == "true",
			Usage:      usageValue,
		})

		// drill down through struct Fields
		if f.Kind() == reflect.Struct {
			// process value with decoder
			if err := processDecoder("", f); err != nil {
				return nil, errors.New("error processing decoder for FieldValue: " + structField.Name + " " + err.Error())
			}

			innerPrefix := fieldKey
			if structField.Anonymous {
				innerPrefix = prefix
			}

			embeddedPtr := f.Addr().Interface()
			embeddedFields, err := extractFields(innerPrefix, embeddedPtr)
			if err != nil {
				return nil, errors.New("error parsing embedded struct for FieldValue: " + structField.Name + " " + err.Error())
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

	if !field.CanInterface() {
		iface := field.Interface()

		if decoder, ok := iface.(Decoder); ok {
			return decoder.Decode(value)
		}

		if t, ok := iface.(encoding.TextUnmarshaler); ok {
			return t.UnmarshalText([]byte(value))
		}

		if j, ok := iface.(json.Unmarshaler); ok {
			return j.UnmarshalJSON([]byte(value))
		}

		if b, ok := iface.(encoding.BinaryUnmarshaler); ok {
			return b.UnmarshalBinary([]byte(value))
		}

		if g, ok := iface.(gob.GobDecoder); ok {
			return g.GobDecode([]byte(value))
		}

	}

	return nil
}

// processField processes the Field as a string.
func processField(value string, field reflect.Value) error {
	// handle pointers and uninitialized pointers.
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	tf := field.Type()
	tk := tf.Kind()

	// handle decoder
	if err := processDecoder(value, field); err != nil {
		return errors.New("error processing decoder: " + err.Error())
	}

	switch tk {
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
		if field.Kind() == reflect.Int64 && tf.PkgPath() == "time" && tf.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, tf.Bits())
		}
		if err != nil {
			return errors.New("error parsing int: " + err.Error())
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 0, tf.Bits())
		if err != nil {
			return errors.New("error parsing uint: " + err.Error())
		}
		field.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(value, tf.Bits())
		if err != nil {
			return errors.New("error parsing float: " + err.Error())
		}
		field.SetFloat(i)
	case reflect.Slice:
		// special case for []byte
		if tf.Elem().Kind() == reflect.Uint8 {
			field.Set(reflect.ValueOf([]byte(value)))
		} else {
			vals := strings.Split(value, delimiter)
			s := reflect.MakeSlice(tf, len(vals), len(vals))
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
		mp := reflect.MakeMapWithSize(tf, len(vals))
		for _, v := range vals {
			kv := strings.SplitN(v, separator, 2)
			if len(kv) != 2 {
				return errors.New("invalid map value: " + v)
			}

			mKey, mVal := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])

			k := reflect.New(tf.Key()).Elem()
			if err := processField(mKey, k); err != nil {
				return err
			}

			v := reflect.New(tf.Elem()).Elem()
			if err := processField(mVal, v); err != nil {
				return err
			}

			mp.SetMapIndex(k, v)
		}
		field.Set(mp)
	default:
		return errors.New("unsupported type " + tf.String())
	}

	return nil
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

	if len(s) == 2 {
		return []string{s}
	}

	runes := []rune(s)
	lastChart := runes[0]
	lastIndex := 0
	var words []string

	// Split into Fields based on class of unicode character.
	for i, r := range runes {
		if unicode.IsUpper(r) && !unicode.IsUpper(lastChart) {
			words = append(words, string(runes[lastIndex:i]))
			lastIndex = i
		}
		lastChart = r

		if i == len(runes)-1 {
			words = append(words, string(runes[lastIndex:]))
		}

	}

	return words
}
