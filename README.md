[![Go](https://github.com/farrukhny/config/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/farrukhny/config/actions/workflows/go.yml)

# Config

Package config implements decoding environment variables and command line flags.
Package idiom is to use a struct to define the configuration options. 
The struct tags are used to define the environment variable and command line flag names and other options.

### Supported struct tags

- `env` - name of environment variable
- `flag` - name of command line flag
- `shortFlag` - short name of command line flag if needed
- `default` - default value for struct field
- `required` - if true, struct field must be provided
- `mask` - if true, value of struct field will be masked in console print with `*` character
- `usage` - usage description for struct field (used in help message)

#### Example

```go
package main

import (
	"fmt"
	"os"

	"github.com/farrukhny/config"
)

type Config struct {
	Host     string `env:"HOST" flag:"host" default:"localhost" usage:"host name"`
	Port     int    `env:"PORT" flag:"port" default:"8080" usage:"port number"`
	User     string `env:"USER" flag:"user" required:"true" usage:"user name"`
	Password string `env:"PASSWORD" flag:"password" required:"true" mask:"true" usage:"user's password"`
}

func main() {
	v = config.Version{
		Major: 1,
		Minor: 0,
		Patch: 0,
		PreRelease: "alpha",
	}
	
	cfg := &Config{}
	err := config.Process("prefix", cfg)	
	if err != nil {
		switch err {
		case config.ErrHelp:
			fmt.Println(config.PrintHelp(&c))
			return
		case config.ErrVersion:
			fmt.Println(v.Version())
			return
		default:
			fmt.Println(err)
			return
		}
	}

	fmt.Printf("%+v\n", cfg)
}
```

_**Command line flag will overwrite value from any source, it will be the highest priority._**


### Environment Variables

`env` tag is used to specify environment variables name. If `env` tag is not specified, the struct key name will be used
as environment variables name.
Struct key name will be converted from camel case to upper case with underscore. For example, `FooBar` will be converted
to `FOO_BAR`.
As a result, the environment variables name will be `FOO_BAR`.

```go
type Config struct {
    FooBar string `env:"FOO_BAR"`
}

func main() {
	
	v = config.Version{
		Major: 1,
		Minor: 0,
		Patch: 0,
		PreRelease: "alpha",
	}
	
    os.Setenv("FOO_BAR", "foo")
    c := Config{}
    err := config.Process(&c)
	if err != nil {
        switch err {
        case config.ErrHelp:
			fmt.Println(config.PrintHelp(&c))
			return
		case config.ErrVersion:
			fmt.Println(v.Version()) // v1.0.0-alpha
			return
        default:
			fmt.Println(err)
            return
        }
	}
	
    fmt.Println(c.FooBar) // foo
}
```


### Command Line Flags

`flag` tag is used to specify command line flag name. If `flag` tag is not specified, the struct key name will be used
as command line flag name.
Struct key name will be converted from camel case to lower case with dash. For example, `FooBar` will be converted
to `foo-bar`.
Command line flags format will be long flag format, for example, `--foo-bar`. If `shortFlag` tag is specified, short flag format
will be used, for example, `-f`.

```go

type Config struct {
    FooBar string `flag:"foo-bar" shortFlag:"f"`
}

func main() {
    os.Args = []string{"", "--foo-bar", "foo"}
    c := Config{}
    config.Process(&c)
    fmt.Println(c.FooBar) // foo
}

```


### Default Value

`default` tag is used to specify default value. If `default` tag is not specified, the default value will be nil.

```go
type Config struct {
    FooBar string `default:"foo"`
}

func main() {
    c := Config{}
    config.Process(&c)
    fmt.Println(c.FooBar) // foo
}
```


### Required Value

`required` tag is used to specify required value. By default, all values are optional. 
If `required` tag is specified, the value must be provided. If value is not provided, the error will be returned.

```go
type Config struct {
    FooBar string `required:"true"`
}

func main() {
    c := Config{}
    err := config.Process(&c)
    fmt.Println(err) // required key FOO_BAR is not set
}
```


### Default Value and Required Value Conflict

Both `default` and `required` tags can not be specified for the same struct field. Error will be returned if both tags are specified.

```go
type Config struct {
    FooBar string `default:"foo" required:"true"`
}

func main() {
    c := Config{}
    err := config.Process(&c)
    fmt.Println(err) // default value and required value can't be specified at the same time
}
```


### Mask Value

`mask` tag is used to specify sensitive value. If `mask` tag is specified, the value will be masked in console print with `*` character.

```go
type Config struct {
    FooBar string `mask:"true"`
}

func main() {
    c := Config{}
    err := config.Process(&c)
    if err != nil {
        switch err {
        case config.ErrHelp:
            fmt.Println(config.PrintHelp(&c))
            return
        case config.ErrVersion:
            fmt.Println(v.Version())
            return
        default:
        fmt.Println(err)
        return
		}
    }
	
	fmt.Println(config.String(&c)) // FooBar: ***
}
```


### Usage

`Usage` tag is used to specify usage of config.

```go
type Config struct {
    FooBar string `usage:"foo bar"`
}

func main() {
    c := Config{}
    config.Process(&c)
    fmt.Println(c.FooBar) // foo bar
}
```


### Decoder interface

`Decoder` interface is used to implement custom decoder. `Decoder` interface has one method `Decode` which accepts value and returns `error`.
Decoder can be used to decode custom types.

```go
type Decoder interface {
    Decode(value string) error
}
```

#### Example

```go

type MyType struct {
    Value string
}

func (m *MyType) Decode(value string) error {
    m.Value = value
    return nil
}

type Config struct {
    FooBar MyType `env:"FOO_BAR"`
}

func main() {
    os.Setenv("FOO_BAR", "foo")
    c := Config{}
    config.Process(&c)
    fmt.Println(c.FooBar.Value) // foo
}

```


### Parser interface
`Parser` interface is used to implement custom parser. `Parser` interface has one method `Parse` which accepts `interface{}` and  return `error`.
Parser can be used to implement unmarshalling for different types. For example, YAML, TOML, JSON, etc.

```go
type Parser interface {
    Parse(v interface{}) error
}
```

#### Example

```go
var yamlData = `
fooBar: foo
`

type YamlConfig struct {
    FooBar string `yaml:"fooBar"`
}

func func main() {
    c := YamlConfig{}
    config.ProcessWithParser("prefix", &c, yaml.WithData([]byte(yamlData)))
    fmt.Println(c.FooBar) // foo
}

```


### Mutator of value

`MutatorFunc` is used to mutate value. `MutatorFunc` is a function which accepts key, value and returns mutated value and error.
Mutator usefully when you need to mutate value before it will be processed. For example, you can get secret path and pull value from secret storage.

Embedded struct can be pointed via KeyPath. For example, `Foo.Bar` will be converted to `FOO_BAR`.

```go
type MutatorFunc func(key string, value string) (string, error)
```

#### Example

```go
type Config struct {
    FooBar string `env:"FOO_BAR"`
}

func main() {
    os.Setenv("FOO_BAR", "foo")
    c := Config{}
    err := config.Process("prefix", &c, config.MutatorFunc(func(key string, value string) (string, error) {
        if key == "FOO_BAR" {
            return "bar", nil
        }
        return value, nil
    }))
    if err != nil {
        switch err {
        case config.ErrHelp:
            fmt.Println(config.PrintHelp(&c))
            return
        case config.ErrVersion:
            fmt.Println(v.Version())
            return
        default:
        fmt.Println(err)
        return
		}
    fmt.Println(c.FooBar) // bar
}
```



