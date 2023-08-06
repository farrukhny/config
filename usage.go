package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
	"text/template"
)

// usageTemplate is the template for the usage message.
var usageTemplate = `Usage: {{.AppName}} [options] [arguments]

{{if .Description}}{{.Description}}{{end}}

Options:
{{range .Field }}
{{- if .ShortFlag }}
	{{- printf "\t-%c," .ShortFlag }}
{{- else}}
	{{- printf "\t " }}
{{- end }}
{{- if .Flag }}
	{{- printf "\t--%s | $%s %s" .Flag .EnvVar (formatFieldType .FieldValue) }}
{{- end }}
	{{- printf "\t%s" (formatField .Default .Usage .Required) }}
{{ end }}
Global Options:
	{{ printf "\t -h," }}{{ printf "\t--help" }}{{ printf "\tshow this help message" }}
	{{ printf "\t -v," }}{{ printf "\t--version" }}{{ printf "\tshow version" }}
`

// GenerateUsageMessage generates the usage message.
func GenerateUsageMessage(cfg interface{}) (string, error) {
	usage, err := extractFields(nil, cfg)
	if err != nil {
		return "", err
	}

	funcMap := template.FuncMap{
		"formatFieldType": formatFieldType,
		"formatField":     formatField,
	}

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', tabwriter.TabIndent)

	err = template.Must(template.New("usage").Funcs(funcMap).Parse(usageTemplate)).Execute(w, struct {
		AppName     string
		Description string
		Field       []Field
	}{
		AppName:     os.Args[0],
		Description: "Configure the application using environment variables and command line flags. See options below.",
		Field:       usage,
	})
	if err != nil {
		return "", err
	}

	err = w.Flush()
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// GenerateStartupMessage generates the startup message.
func GenerateStartupMessage(cfg interface{}) (string, error) {
	cfgUsage, err := extractFields(nil, cfg)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s is starting up with the following configuration:\n", os.Args[0]))
	for _, f := range cfgUsage {
		val := valueToString(f.FieldValue)
		sb.WriteString(fmt.Sprintf("--	%s: %v\n", f.Flag, maskString(val, f.Mask)))

	}

	return sb.String(), nil
}

// GenerateJSONStartupMessage generates the startup message in JSON format.
func GenerateJSONStartupMessage(cfg interface{}) (string, error) {
	cfgUsage, err := extractFields(nil, cfg)
	if err != nil {
		return "", err
	}

	startupMessage := make(map[string]interface{})
	for _, f := range cfgUsage {
		startupMessage[f.Flag] = maskString(valueToString(f.FieldValue), f.Mask)
	}

	jsonMsg, err := json.Marshal(startupMessage)
	if err != nil {
		return "", err
	}

	return string(jsonMsg), nil
}

// maskString masks the string if the mask is set to true.
func maskString(s string, mask bool) string {
	if mask && len(s) > 3 {
		return strings.Repeat("*", len(s)-3) + s[len(s)-3:]
	}
	return s
}

// formatField formats the field information into a single string.
func formatField(defaultValue, usage string, required bool) string {
	var value string
	if required {
		value = "(required)"
	}

	if defaultValue != "" {
		value = fmt.Sprintf("(default: %s)", defaultValue)
	}

	if usage != "" {
		return fmt.Sprintf("%s %s", usage, value)
	}

	return value
}

// formatFieldType formats the field type into a single human-readable string.
func formatFieldType(f reflect.Value) string {
	// check if field is time.Duration type and format accordingly
	if f.Type().String() == "time.Duration" {
		return "duration"
	}

	return f.Type().String()
}
