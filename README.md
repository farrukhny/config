# Config

[![Go](https://github.com/farrukhny/config/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/farrukhny/config/actions/workflows/go.yml)
[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/farrukhny/config)](https://github.com/farrukhny/config)
[![Go Report Card](https://goreportcard.com/badge/github.com/farrukhny/config)](https://goreportcard.com/report/github.com/farrukhny/config)

The config package provides functionality for loading and processing configuration values from environment variables and command line flags. It offers support for custom decoders, parsers, and mutators.

To install the config package, use the following command:

```bash
go get github.com/farrukhny/config
```

## Usage

#### Tags

The `config` package uses struct tags to specify configuration options. Here are the available tags:


- `env`: Specifies the environment variable name for the field.
- `default`: Specifies the default value for the field.
- `required`: Specifies whether the field is required. Default value can not be used if the field is required.
- `usage`: Specifies the description of the field.
- `flag`: Specifies the command line flag name for the field.
- `shortFlag`: Specifies the short command line flag name for the field.
- `mask`: Specifies whether the field value should be masked in the output.


### Defining Configuration Struct

First, define a struct that represents your application's configuration settings. You can use struct tags to specify environment variable names, default values, and more.
```go
type AppConfig struct {
    Port     int    `env:"APP_PORT" default:"8080" usage:"Port number"`
    LogLevel string `env:"LOG_LEVEL" default:"info" usage:"Log level"`
    // ... other configuration fields
}
```

### Loading Configuration

Load the configuration using the `config.Process` function. It will read environment variables and command-line flags, applying any specified mutators.

#### Basic Usage:

```go
func Process(cfg interface{}, mutator ...MutatorFunc) error
```

```go
func main() {
    var cfg AppConfig
    err := config.Process(&cfg)
    if err != nil {
        // Handle error
    }

    // Your application logic using cfg
}
```

### Using Parsers

You can also use custom parsers to load configuration from different sources, such as files or remote services. Create parsers that implement the `config.Parser` interface.

```go
type Parser interface {
    Parse(cfg interface{}) error
}
```
```go
type FileParser struct {
    FilePath string
}

func (p *FileParser) Parse(cfg interface{}) error {
    // Read configuration from file and populate cfg
    return nil
}
```

Then, use the `config.ProcessWithParser` function:

```go
func ProcessWithParser(cfg interface{}, parsers []Parser, mutator ...MutatorFunc) error
```

```go
func main() {
    var cfg AppConfig
    err := config.ProcessWithParser(&cfg, []config.Parser{&FileParser{FilePath: "config.json"}})
    if err != nil {
        // Handle error
    }

    // Your application logic using cfg
}
```

### Custom Decoders

The `Decoder` interface declares the Decode method, which can be implemented to provide custom decoding logic.

```go
type Decoder interface {
    Decode(val string) error
}
```

#### Example Usage

```go
import (
    "encoding/base64"
)

type Base64Decoder struct{}

func (d *Base64Decoder) Decode(val string) error {
    decoded, err := base64.StdEncoding.DecodeString(val)
    if err != nil {
        return err
    }
    
    // Use the decoded value
    // e.g., return the decoded value or set it to a field

    return nil
}

```

### Mutator Functions

The `MutatorFunc` is a function type that mutates a value of a key before it is set to the field.
for example, pulling secrets from a secret manager and setting them to the configuration struct.

```go
type MutatorFunc func(key, value string) (string, error)
```

#### Example Usage

```go
func secretMutator(key, value string) (string, error) {
// Retrieve secret using the key and any necessary logic
secretValue, err := retrieveSecretFromManager(key)
    if err != nil {
        return "", err
    }
    return secretValue, nil
}
```
