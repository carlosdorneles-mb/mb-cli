package ports

import "context"

// Logger is the minimal logging surface used by use cases.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any) error
	Warn(ctx context.Context, msg string, args ...any) error
	Debug(ctx context.Context, msg string, args ...any) error
	Error(ctx context.Context, msg string, args ...any) error
}
