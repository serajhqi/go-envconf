# envconf

Lightweight Go library to load typed config structs from environment variables and `.env` files.

## Install

```bash
go get github.com/serajhqi/envconf
```

## Quick Start

```go
package main

import (
	"fmt"
	"time"

	"github.com/serajhqi/envconf"
)

type Config struct {
	DBPass  string        `env:"DB_PASS" default:"postgres"`
	Debug   bool          `env:"DEBUG" default:"true"`
	Port    int           `env:"PORT" default:"8080"`
	Timeout time.Duration `env:"TIMEOUT" default:"5s"`
}

func main() {
	_, err := envconf.Load[Config]()
	if err != nil {
		panic(err)
	}

	cfg, err := envconf.Get[Config]()
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Port, cfg.Debug)
}
```

## Tags

- `env:"KEY"` sets the environment variable name.
- `default:"value"` sets fallback value.

## Value Priority

1. OS environment variable
2. `.env` file
3. `default` tag

## Options

- `envconf.WithDotEnvPath("/path/to/.env")`
- `envconf.WithoutDotEnv()`

## Singleton Lifecycle

1. Call `Load[T](...)` once to initialize config.
2. Call `Get[T]()` anywhere to access the same loaded instance.
3. `MustLoad[T](...)` and `MustGet[T]()` are panic-on-error variants.

## Supported Types

- `string`
- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`
- `float32`, `float64`
- `time.Duration`
- Named/alias types of the above scalar kinds
