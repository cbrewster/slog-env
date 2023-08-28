# slog-env

[![Go Reference](https://pkg.go.dev/badge/github.com/cbrewster/slog-env.svg)](https://pkg.go.dev/github.com/cbrewster/slog-env)

slog-env provides a [log/slog](https://pkg.go.dev/log/slog) handler which allows setting the log level
via the GO_LOG environment variable. Additionally it allows setting the log level on a per-package basis.

Examples:
  - `GO_LOG=info` will set the log level to info globally.
  - `GO_LOG=info,mypackage=debug` will set the log level to info by default, but sets it to debug for logs from mypackage.
  - `GO_LOG=info,mypackage=debug,otherpackage=error` you can specify multiple packages by using a comma separator.

## Installation

```
go get github.com/cbrewster/slog-env
```

## Usage

To set up `slog-env`, wrap your normal slog handler:

```go
import (
    slogenv "github.com/cbrewster/slog-env"
)

func main() {
    logger := slog.New(slogenv.NewHandler(slog.NewTextHandler(os.Stderr, nil)))
    // ...
}
```

```bash
$ GO_LOG=info,mypackage=debug go run .
```
