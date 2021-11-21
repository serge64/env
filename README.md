# env

Package env provides an `env` struct field tag to unmarshal environment variables.

Supported type type or unmarshaling:
* int, int8, int16, int32, int64
* float32, float64
* time.Duration
* string
* bool

## Example of use

```go
package main

import (
    "log"

    "github.com/serge64/env"
)

type Config struct {
    Duration     time.Duration `env:"TYPE_DURATION"`
    DefaultValue string        `env:"MISSING_ENV,default=default_value"`
    Bool         bool          `env:"TYPE_BOOL"`
}

func main() {
    var config Config
    if err := env.Unmarshal(&config); err != nil {
        log.Fatal(err)
    }
}
```
