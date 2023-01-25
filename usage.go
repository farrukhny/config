package config

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
)

// usageTemplate is the template for the usage message.
var usageTemplate = `Usage: {{.AppName}} [options] [arguments]

{{if .Description}}{{.Description}}{{end}}

Options:
{{- range .Field }}
{{ if .ShortFlag }}{{ printf "\t -%c," .ShortFlag }}{{else}}{{ printf "\t " }}{{ end }}{{ if .Flag }}{{ printf "\t--%s" .Flag }}{{ end }}{{ if .EnvVar }}{{ printf "\t($%s)" .EnvVar }}{{else}}{{ printf "\t "}}{{ end }}{{ printf "\t%s" .FieldType }}{{ if .Required }}{{ if .Usage }}{{ printf "\t%s (required)" .Usage }}{{else}}{{ printf "\t(required)" }}{{ end }}{{else if .Default}}{{ if .Usage }}{{ printf "\t%s (default: %s)" .Usage .Default }}{{else}}{{ printf "\t(default: %s)" .Default }}{{ end }}{{else}}{{ printf "\t%s" .Usage }}{{end}}{{ end }}

Global Options:
{{ printf "\t -h," }}{{ printf "\t--help" }}{{ printf "\tshow this help message" }}
{{ printf "\t -v," }}{{ printf "\t--version" }}{{ printf "\tshow version" -}}
`

// PrintUsage will print the usage message. It will accept a conf struct and run extractFields on it.
// It will then use Filed struct from extractFields to build the usage message with the usageTemplate.
func PrintUsage(cfg interface{}) (string, error) {
	usage, err := extractFields(nil, cfg)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	tmp, err := template.New("usage").Parse(usageTemplate)
	if err != nil {
		return "", err
	}

	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', tabwriter.TabIndent)
	defaultDescription := `This application is configured via the environment variables or command line flags.`

	// if the conf struct has masked it will mask the default value
	var fields []Field
	for _, f := range usage {
		f.Default = maskString(f.Default, f.Mask)

		fields = append(fields, f)
	}

	err = tmp.Execute(w, struct {
		AppName     string
		Description string
		Field       []Field
	}{
		AppName:     os.Args[0],
		Description: defaultDescription,
		Field:       fields,
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

// PrintStartupMessage will print the startup message. It will accept a conf struct and run extractFields on it.
// It will then use Filed struct from extractFields to build the startup message.
func PrintStartupMessage(cfg interface{}) (string, error) {
	cfgUsage, err := extractFields(nil, cfg)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s is starting up with the following configuration:\n", os.Args[0]))
	for _, f := range cfgUsage {
		sb.WriteString(fmt.Sprintf("--%s: %s\n", f.Flag, maskString(f.Default, f.Mask)))
	}

	return sb.String(), nil
}

// maskString will mask the string if the mask is set to true.
func maskString(s string, mask bool) string {
	// check if string is empty
	if s == "" {
		return ""
	}

	// check if mask is set
	if mask {
		msk := strings.Repeat("*", len(s))
		return strings.Replace(s, s, msk, 1)
	}

	return s
}
