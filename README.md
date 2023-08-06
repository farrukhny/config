# Config

[![Go](https://github.com/farrukhny/config/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/farrukhny/config/actions/workflows/go.yml)


The config package provides functionality for loading and processing configuration values from environment variables and command line flags. It offers support for custom decoders, parsers, and mutators.

To install the config package, use the following command:

```bash
go get github.com/example/config
```

## Usage

### Tags

The `config` package uses struct tags to specify configuration options. Here are the available tags:


- `env`: Specifies the environment variable name for the field.
- `default`: Specifies the default value for the field.
- `required`: Specifies whether the field is required. Default value can not be used if the field is required.
- `usage`: Specifies the description of the field.
- `flag`: Specifies the command line flag name for the field.
- `shortFlag`: Specifies the short command line flag name for the field.
- `mask`: Specifies whether the field value should be masked in the output.


### Processing Configuration

The `Process` function processes a struct with environment variables and command line flags as the source. It accepts a prefix string that is used as a prefix for environment variables, a cfg interface{} representing the configuration struct, and optional mutator functions that mutate the value before setting it to the field.

```go
func Process(cfg interface{}, mutator ...MutatorFunc) error
```

The `ProcessWithParser` function processes the struct with the given parsers, followed by environment variables and command line flags as the source. It accepts a prefix string, a cfg interface{} representing the configuration struct, an array of parsers, and optional mutator functions.

```go
func ProcessWithParser(cfg interface{}, parsers []Parser, mutator ...MutatorFunc) error
```

### Custom Decoders

The `Decoder` interface declares the Decode method, which can be implemented to provide custom decoding logic.

```go
type Decoder interface {
    Decode(val string) error
}
```

### Parsers

The `Parser` interface declares the `Parse` method, which can be implemented to extend the functionality of the parsers used to unmarshal the config.

```go
type Parser interface {
    Parse(cfg interface{}) error
}
```

### Source

The `source` interface declares the `Source` method, which is used to load the configuration from environment variables and command line flags. The Source method accepts a `Field` struct.

```go
type source interface {
    Source(f Field) (string, bool)
}
```

### Mutator Functions

The `MutatorFunc` is a function type that mutates a value of a key before it is set to the field.

```go
type MutatorFunc func(key, value string) (string, error)
```

### Example Usage

```go

package main

import (
	"fmt"

	"github.com/example/config"
)

type Config struct {
	Host     string `env:"HOST" flag:"host" default:"localhost" usage:"Server host"`
	Port     int    `env:"PORT" flag:"port" default:"8080" usage:"Server port"`
	Username string `env:"USERNAME" flag:"username" required:"true" usage:"Username"`
	Password string `env:"PASSWORD" flag:"password" mask:"true" usage:"Password"`
}

func main() {
	var cfg Config

	err := config.Process(&cfg)
	if err != nil {
		fmt.Println("Error processing configuration:", err)
		return
	}

	fmt.Println("Host:", cfg.Host)
	fmt.Println("Port:", cfg.Port)
	fmt.Println("Username:", cfg.Username)
	fmt.Println("Password:", cfg.Password)
}

```

