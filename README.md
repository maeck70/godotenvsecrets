# godotenvsecrets

A Go package for loading environment variables and secrets from various providers, including AWS and Azure, as well as standard environment variables. Designed for use in cloud-native applications and local development.

## Features

- Load environment variables from `.env` files
- Retrieve secrets from AWS Secret Manager
- Support for Azure (not yet implemented)
- Fallback to standard environment variables
- Table-driven tests for robust validation

## Usage

### Installation

Add to your Go module:

```bash
go get github.com/maeck70/godotenvsecrets
```

### Example

```go
package main

import (
    "fmt"
    "github.com/maeck70/godotenvsecrets"
)

func main() {
    err := godotenvsecrets.Load()
    if err != nil {
        fmt.Println("Error loading .env:", err)
    }
    secret, err := godotenvsecrets.Getenv("@aws:dev/goenvsecrets:serviceaccount")
    if err != nil {
        fmt.Println("Error getting secret:", err)
    } else {
        fmt.Println("Service Account:", secret)
    }
}
```

## Testing

Run all tests:

```bash
go test
```

## API

- `Load() error`: Loads environment variables from `.env` file.
- `Getenv(key string) (string, error)`: Retrieves the value for a given key, supporting secret providers.

## License

MIT
