package config_test

import (
	"os"
	"testing"

	"github.com/farrukhny/config"
	"github.com/google/go-cmp/cmp"
)

const (
	success = "\u2713"
	failed  = "\u2717"
)

// conf is the struct that will be used to test the conf package.
type conf struct {
	Host string `yaml:"host" env:"HOST" flag:"host" default:"localhost"`
	Port int    `yaml:"port" env:"PORT" default:"8080"`
	DB   string `yaml:"db" env:"DB" default:"postgres"`
	HTTP HTTP   `yaml:"http" env:"HTTP"`
	Embedded
}

type HTTP struct {
	Host string `yaml:"host" env:"HTTP_HOST" default:"localhost" mask:"true"`
}

type Embedded struct {
	ApiKey    string `env:"API_KEY" default:"1234567890"`
	ApiUrl    string `env:"API_URL" default:"https://example.com"`
	ApiSecret string `env:"API_SECRET" mask:"true"`
}

var mutateValue config.MutatorFunc = func(key string, value string) (string, error) {
	if key == "Host" {
		return "mutated-host", nil
	}

	if key == "HTTP_Host" {
		return "mutated-http-host", nil
	}

	return value, nil
}

func TestProcess(t *testing.T) {
	test := []struct {
		name    string
		envs    map[string]string
		args    []string
		mutator []config.MutatorFunc
		want    conf
	}{
		{
			name:    "Default",
			envs:    map[string]string{},
			args:    []string{},
			mutator: nil,
			want: conf{
				Host: "localhost",
				Port: 8080,
				DB:   "postgres",
				HTTP: HTTP{
					Host: "localhost",
				},
				Embedded: Embedded{
					ApiKey: "1234567890",
					ApiUrl: "https://example.com",
				},
			},
		},
		{
			name: "Envs",
			envs: map[string]string{
				"HOST":       "http",
				"PORT":       "8081",
				"DB":         "db",
				"HTTP_HOST":  "http-host",
				"API_KEY":    "api-key",
				"API_URL":    "api-url",
				"API_SECRET": "someSecret",
			},
			args:    []string{},
			mutator: nil,
			want: conf{
				Host: "http",
				Port: 8081,
				DB:   "db",
				HTTP: HTTP{
					Host: "http-host",
				},
				Embedded: Embedded{
					ApiKey:    "api-key",
					ApiUrl:    "api-url",
					ApiSecret: "someSecret",
				},
			},
		},
		{
			name:    "Flags",
			envs:    map[string]string{},
			args:    []string{"conf.test", "--host", "http", "--port", "8081", "--db", "db", "--http-host", "http-host", "--api-key", "api", "--api-url", "api-url", "--api-secret", "someSecret"},
			mutator: nil,
			want: conf{
				Host: "http",
				Port: 8081,
				DB:   "db",
				HTTP: HTTP{
					Host: "http-host",
				},
				Embedded: Embedded{
					ApiKey:    "api",
					ApiUrl:    "api-url",
					ApiSecret: "someSecret",
				},
			},
		},
		{
			name: "Envs and Flags",
			envs: map[string]string{
				"HOST":            "http-env",
				"PORT":            "8080",
				"DB":              "db-env",
				"HTTP_HOST":       "http-host-env",
				"API_KEY":         "api-key-env",
				"TEST_API_URL":    "api-url-env",
				"TEST_API_SECRET": "SomeSecret-Env",
			},
			args:    []string{"conf.test", "--host", "host-flag", "--port", "8081", "--http-host", "http", "--api-key", "api", "--api-url", "api", "--api-secret", "someSecret"},
			mutator: nil,
			want: conf{
				Host: "host-flag",
				Port: 8081,
				DB:   "db-env",
				HTTP: HTTP{
					Host: "http",
				},
				Embedded: Embedded{
					ApiKey:    "api",
					ApiUrl:    "api",
					ApiSecret: "someSecret",
				},
			},
		},
		{
			name: "Mutator",
			envs: map[string]string{
				"HOST":       "http-env",
				"PORT":       "8080",
				"DB":         "db-env",
				"HTTP_HOST":  "http-host-env",
				"API_KEY":    "api-key",
				"API_URL":    "api-url",
				"API_SECRET": "someSecret",
			},
			args:    nil,
			mutator: []config.MutatorFunc{mutateValue},
			want: conf{
				Host: "mutated-host",
				Port: 8080,
				DB:   "db-env",
				HTTP: HTTP{
					Host: "mutated-http-host",
				},
				Embedded: Embedded{
					ApiKey:    "api-key",
					ApiUrl:    "api-url",
					ApiSecret: "someSecret",
				},
			},
		},
	}

	for _, tt := range test {
		t.Logf("Given the need to test the Process function with %s", tt.name)
		{

			os.Clearenv()
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}

			f := func(t *testing.T) {
				os.Args = tt.args
				var cfg conf
				if err := config.Process(&cfg, tt.mutator...); err != nil {
					t.Fatalf("\t%s\tShould be able to process the conf struct: %v", failed, err)
				}
				t.Logf("\t%s\tShould be able to process the conf struct.", success)

				if diff := cmp.Diff(tt.want, cfg); diff != "" {
					t.Fatalf("\t%s\tShould get the expected config: %s", failed, diff)
				}
				t.Logf("\t%s\tShould get the expected config.", success)
			}
			t.Run(tt.name, f)
		}
	}
}
