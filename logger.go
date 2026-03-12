package modbus

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
)

// Logger is the logging interface accepted by ClientConfiguration and ServerConfiguration.
// Implement this interface to integrate any structured or levelled logging library
// (e.g. zap, zerolog, slog, logrus).
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// NewStdLogger wraps a stdlib *log.Logger so it satisfies the Logger interface.
// If l is nil, output is written to os.Stdout with no flags.
func NewStdLogger(l *log.Logger) Logger {
	if l == nil {
		l = log.New(os.Stdout, "", 0)
	}

	return &stdLogger{l: l}
}

type stdLogger struct{ l *log.Logger }

func (sl *stdLogger) Debugf(format string, args ...any) { sl.l.Printf(format, args...) }
func (sl *stdLogger) Infof(format string, args ...any)  { sl.l.Printf(format, args...) }
func (sl *stdLogger) Warnf(format string, args ...any)  { sl.l.Printf(format, args...) }
func (sl *stdLogger) Errorf(format string, args ...any) { sl.l.Printf(format, args...) }

// NewSlogLogger wraps a slog.Handler so it satisfies the Logger interface.
// Use slog.NewJSONHandler, slog.NewTextHandler, or any third-party handler.
func NewSlogLogger(h slog.Handler) Logger {
	return &slogLogger{sl: slog.New(h)}
}

type slogLogger struct{ sl *slog.Logger }

func (ll *slogLogger) Debugf(format string, args ...any) {
	ll.sl.DebugContext(context.Background(), fmt.Sprintf(format, args...))
}

func (ll *slogLogger) Infof(format string, args ...any) {
	ll.sl.InfoContext(context.Background(), fmt.Sprintf(format, args...))
}

func (ll *slogLogger) Warnf(format string, args ...any) {
	ll.sl.WarnContext(context.Background(), fmt.Sprintf(format, args...))
}

func (ll *slogLogger) Errorf(format string, args ...any) {
	ll.sl.ErrorContext(context.Background(), fmt.Sprintf(format, args...))
}

// NopLogger returns a Logger that discards all log output.
// Useful in tests or when logging is intentionally disabled.
func NopLogger() Logger {
	return &nopLogger{}
}

type nopLogger struct{}

func (nl *nopLogger) Debugf(string, ...any) {}
func (nl *nopLogger) Infof(string, ...any)  {}
func (nl *nopLogger) Warnf(string, ...any)  {}
func (nl *nopLogger) Errorf(string, ...any) {}

// logger is the internal prefixing adapter used by transports, client, and server.
// It wraps a public Logger, prepending the modbus component prefix and level tag
// to every message so call sites remain simple.
type logger struct {
	prefix string
	inner  Logger
}

// newLogger creates a prefixing logger. If l is nil the slog default logger is used
// (Go 1.21+). Configure slog.SetDefault to change the global default.
func newLogger(prefix string, l Logger) *logger {
	if l == nil {
		l = NewSlogLogger(slog.Default().Handler())
	}

	return &logger{prefix: prefix, inner: l}
}

func (l *logger) Debug(msg string) {
	l.inner.Debugf("%s [debug]: %s", l.prefix, msg)
}

func (l *logger) Debugf(format string, args ...any) {
	l.inner.Debugf("%s [debug]: "+format, logPrepend(l.prefix, args)...)
}

func (l *logger) Info(msg string) {
	l.inner.Infof("%s [info]: %s", l.prefix, msg)
}

func (l *logger) Infof(format string, args ...any) {
	l.inner.Infof("%s [info]: "+format, logPrepend(l.prefix, args)...)
}

func (l *logger) Warning(msg string) {
	l.inner.Warnf("%s [warn]: %s", l.prefix, msg)
}

func (l *logger) Warningf(format string, args ...any) {
	l.inner.Warnf("%s [warn]: "+format, logPrepend(l.prefix, args)...)
}

func (l *logger) Error(msg string) {
	l.inner.Errorf("%s [error]: %s", l.prefix, msg)
}

func (l *logger) Errorf(format string, args ...any) {
	l.inner.Errorf("%s [error]: "+format, logPrepend(l.prefix, args)...)
}

func (l *logger) Fatal(msg string) {
	l.Error(msg)
}

func (l *logger) Fatalf(format string, args ...any) {
	l.Errorf(format, args...)
}

// logPrepend inserts v at the front of args, returning a new slice.
func logPrepend(v any, args []any) []any {
	return append([]any{v}, args...)
}
