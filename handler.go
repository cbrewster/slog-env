// Package slogenv provides a [log/slog] handler which allows setting the log level
// via the GO_LOG environment variable. Additionally it allows setting the log level on a per-package basis.
//
// Examples:
//   - GO_LOG=info will set the log level to info globally.
//   - GO_LOG=info,mypackage=debug will set the log level to info by default, but sets it to debug for logs from mypackage.
//   - GO_LOG=info,mypackage=debug,otherpackage=error you can specify multiple packages by using a comma separator.
//
// To set up slog-env, wrap your normal slog handler:
//
//		import (
//			slogenv "github.com/cbrewster/slog-env"
//		)
//
//		func main() {
//			logger := slog.New(slogenv.NewHandler(slog.NewTextHandler(os.Stderr, nil)))
//			// ...
//		}
package slogenv

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

type config struct {
	defaultLevel  slog.Level
	envVarName    string
	defaultFilter string
}

// Opt allows customizing the handler's configuration.
type Opt func(*config)

// WithDefaultLevel sets the default log level if the environment variable is not set.
func WithDefaultLevel(level slog.Level) Opt {
	return func(cfg *config) {
		cfg.defaultLevel = level
	}
}

// WithEnvVarName sets the environment variable used to set the log level. Default is GO_LOG.
func WithEnvVarName(name string) Opt {
	return func(cfg *config) {
		cfg.envVarName = name
	}
}

// WithDefaultFilter sets the default filter if the environment variable is not set.
func WithDefaultFilter(filter string) Opt {
	return func(cfg *config) {
		cfg.defaultFilter = filter
	}
}

// Handler is a log handler that dynamically sets the log level based on the GO_LOG environment variable.
// The log level can be set on a per-package basis.
type Handler struct {
	inner slog.Handler
	// defaultLevel is the log level used for logs not matching one of the package filters.
	defaultLevel slog.Level
	// perPackageLevel stores the log level for each package.
	perPackageLevel map[string]slog.Level
}

var _ slog.Handler = (*Handler)(nil)

// NewHandler creates a new env logger handler.
func NewHandler(inner slog.Handler, opts ...Opt) *Handler {
	cfg := config{
		envVarName:   "GO_LOG",
		defaultLevel: slog.LevelInfo,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	filter := cfg.defaultFilter
	if envFilter := os.Getenv(cfg.envVarName); envFilter != "" {
		filter = envFilter
	}

	defaultLevel, perPackageLevel := parseFilter(cfg.defaultLevel, filter)

	return &Handler{
		defaultLevel:    defaultLevel,
		perPackageLevel: perPackageLevel,
		inner:           inner,
	}
}

// Enabled implements slog.Handler.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	if len(h.perPackageLevel) == 0 {
		return level >= h.defaultLevel
	}

	// Unfortunately, when filtering by package, we need to wait
	// until Handle is called before we determine if a log is enabled.
	return true
}

// Handle implements slog.Handler.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	level := h.getLevelForRecord(record)

	if record.Level < level {
		return nil
	}

	return h.inner.Handle(ctx, record)
}

// WithAttrs implements slog.Handler.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		inner:           h.inner.WithAttrs(attrs),
		defaultLevel:    h.defaultLevel,
		perPackageLevel: h.perPackageLevel,
	}
}

// WithGroup implements slog.Handler.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		inner:           h.inner.WithGroup(name),
		defaultLevel:    h.defaultLevel,
		perPackageLevel: h.perPackageLevel,
	}
}

func (h *Handler) getLevelForRecord(record slog.Record) slog.Level {
	if len(h.perPackageLevel) == 0 {
		return h.defaultLevel
	}

	fs := runtime.CallersFrames([]uintptr{record.PC})
	f, _ := fs.Next()
	pkg, ok := parsePackage(f.Function)
	if !ok {
		return h.defaultLevel
	}

	level, ok := h.perPackageLevel[pkg]
	if !ok {
		return h.defaultLevel
	}

	return level
}

// parsePackage parses the package out of a formatted function name.
// Example:
// github.com/cbrewster/slog-env_test.TestFilterPackage
// Will return slog-env_test
func parsePackage(function string) (string, bool) {
	parts := strings.Split(function, "/")
	pkg, _, ok := strings.Cut(parts[len(parts)-1], ".")
	return pkg, ok
}

// parseFilter parses the filter specified in the ENV var.
// The filter can consist of comman separated filters.
// A filter specifies a package and a filter level, if the package is omitted,
// the filter level is used as the default.
//
// This will set the log level to info for all logs
// GO_LOG=info
//
// This will set the log level to error by default, but debug for mypackage and info for otherpackage
// GO_LOG=error,mypackage=debug,otherpackage=info
//
// Filters later in the list have higher precedence over ones earlier in the list.
func parseFilter(defaultLevel slog.Level, filter string) (slog.Level, map[string]slog.Level) {
	perPackageLevel := make(map[string]slog.Level)

	filters := strings.Split(filter, ",")
	for _, filter := range filters {
		first, second, ok := strings.Cut(filter, "=")
		if !ok {
			defaultLevel.UnmarshalText([]byte(first))
			continue
		}

		packageLevel := perPackageLevel[first]
		packageLevel.UnmarshalText([]byte(second))
		perPackageLevel[first] = packageLevel
	}

	return defaultLevel, perPackageLevel
}
