# Log Package

The log package provides structured logging capabilities for Moov applications.

## Usage

### Basic Logging

```go
import "github.com/moov-io/base/log"

// Create a new logger
logger := log.NewDefaultLogger()

// Log a message with different levels
logger.Info().Log("Application started")
logger.Debug().Log("Debug information")
logger.Warn().Log("Warning message")
logger.Error().Log("Error occurred")

// Log with key-value pairs
logger.Info().Set("request_id", log.String("12345")).Log("Processing request")

// Log formatted messages
logger.Infof("Processing request %s", "12345")

// Log errors
err := someFunction()
if err != nil {
    logger.LogError(err)
}
```

### Using Fields

```go
import "github.com/moov-io/base/log"

// Create a map of fields
fields := log.Fields{
    "request_id": log.String("12345"),
    "user_id":    log.Int(42),
    "timestamp":  log.Time(time.Now()),
}

// Log with fields
logger.With(fields).Info().Log("Request processed")
```

### Using StructContext

The `StructContext` function allows you to log struct fields automatically by using tags.

```go
import "github.com/moov-io/base/log"

// Define a struct with log tags
type User struct {
    ID       int    `log:"id"`
    Username string `log:"username"`
    Email    string `log:"email,omitempty"` // won't be logged if empty
    Address  Address `log:"address"` // nested struct must have log tag
    Hidden   string // no log tag, won't be logged
}

type Address struct {
    Street  string `log:"street"`
    City    string `log:"city"`
    Country string `log:"country"`
}

// Create a user
user := User{
    ID:       1,
    Username: "johndoe",
    Email:    "john@example.com",
    Address: Address{
        Street:  "123 Main St",
        City:    "New York",
        Country: "USA",
    },
    Hidden: "secret",
}

// Log with struct context
logger.With(log.StructContext(user)).Info().Log("User logged in")

// Log with struct context and prefix
logger.With(log.StructContext(user, log.WithPrefix("user"))).Info().Log("User details")

// Using custom tag other than "log"
type Product struct {
    ID    int     `otel:"product_id"`
    Name  string  `otel:"product_name"`
    Price float64 `otel:"price,omitempty"`
}

product := Product{
    ID:    42,
    Name:  "Widget",
    Price: 19.99,
}

// Use otel tags instead of log tags
logger.With(log.StructContext(product, log.WithTag("otel"))).Info().Log("Product details")
```

The above will produce log entries with the following fields:
- `id=1`
- `username=johndoe`
- `email=john@example.com`
- `address.street=123 Main St`
- `address.city=New York`
- `address.country=USA`

With the prefix option, the fields will be:
- `user.id=1`
- `user.username=johndoe`
- `user.email=john@example.com`
- `user.address.street=123 Main St`
- `user.address.city=New York`
- `user.address.country=USA`

With the custom tag option, the fields will be extracted from the tag you specify (such as `otel`):
- `product_id=42`
- `product_name=Widget`
- `price=19.99`

Note that nested structs or pointers to structs must have the specified tag to be included in the context.

## Features

- Structured logging with key-value pairs
- Multiple log levels (Debug, Info, Warn, Error, Fatal)
- JSON and LogFmt output formats
- Context-based logging
- Automatic struct field logging with StructContext
- Support for various value types (string, int, float, bool, time, etc.)

## Configuration

The default logger format is determined by the `MOOV_LOG_FORMAT` environment variable:
- `json`: JSON format
- `logfmt`: LogFmt format (default)
- `nop` or `noop`: No-op logger that discards all logs
