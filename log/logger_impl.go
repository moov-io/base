package log

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
)

func NewDefaultLogger() Logger {
	return NewLogger(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)))
}

func NewNopLogger() Logger {
	return NewLogger(log.NewNopLogger())
}

func NewJSONLogger() Logger {
	return NewLogger(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)))
}

func NewTestLogger() Logger {
	if testing.Verbose() {
		return NewDefaultLogger()
	}
	return NewNopLogger()
}

func NewBufferLogger() (*strings.Builder, Logger) {
	buffer := strings.Builder{}
	writer := log.NewLogfmtLogger(log.NewSyncWriter(&buffer))
	log := NewLogger(writer)
	return &buffer, log
}

func NewLogger(writer log.Logger) Logger {
	l := &logger{
		writer: writer,
		ctx:    make(map[string]Valuer),
	}

	// Default logs to be info until changed
	return l.Info()
}

var _ Logger = (*logger)(nil)

type logger struct {
	writer log.Logger
	ctx    map[string]Valuer
}

func (l *logger) Set(key string, value Valuer) Logger {
	return l.With(Fields{
		key: value,
	})
}

// With returns a new Logger with the contexts added to its own.
func (l *logger) With(ctxs ...Context) (out Logger) {
	defer func() {
		if r := recover(); r != nil {
			var file string
			var line int
			var ok bool

			// Search the call stack for the first non-Go file
			for i := 1; i < 10; i++ {
				_, file, line, ok = runtime.Caller(i)
				if !ok || !strings.Contains(file, "/src/runtime/") {
					break
				}
			}

			if ok {
				l.writer.Log(
					"level", "error",
					"file", file,
					"line", fmt.Sprintf("%d", line),
					"msg", fmt.Sprintf("recovered from %T - %v", r, r),
				)
			}

			out = l // make the caller whole
			return
		}
	}()

	// Estimation assuming that for each ctxs has at least 1 value.
	combined := make(map[string]Valuer, len(l.ctx)+len(ctxs))

	for k, v := range l.ctx {
		combined[k] = v
	}

	for _, c := range ctxs {
		if c == nil {
			continue
		}

		itemCtx := c.Context()
		for k, v := range itemCtx {
			combined[k] = v
		}
	}

	return &logger{
		writer: l.writer,
		ctx:    combined,
	}
}

func (l *logger) Details() map[string]interface{} {
	m := make(map[string]interface{}, len(l.ctx))
	for k, v := range l.ctx {
		m[k] = v.getValue()
	}
	return m
}

func (l *logger) Debug() Logger {
	return l.With(Debug)
}

func (l *logger) Info() Logger {
	return l.With(Info)
}

func (l *logger) Warn() Logger {
	return l.With(Warn)
}

func (l *logger) Error() Logger {
	return l.With(Error)
}

func (l *logger) Fatal() Logger {
	return l.With(Fatal)
}

func (l *logger) Log(msg string) {
	// Frontload the timestamp and msg
	keyvals := []interface{}{
		"ts", time.Now().UTC().Format(time.RFC3339),
	}
	if msg != "" {
		keyvals = append(keyvals, "msg", msg)
	}

	// Sort the rest of the list so the log lines look similar
	details := l.Details()
	keys := make([]string, 0, len(l.ctx))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Lets add them into the arguments
	for _, k := range keys {
		keyvals = append(keyvals, k, details[k])
	}

	_ = l.writer.Log(keyvals...)
}

func (l *logger) Logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Log(msg)
}

// Send is equivalent to calling Msg("")
func (l *logger) Send() {
	l.Log("")
}

func (l *logger) LogError(err error) LoggedError {
	l.Set("errored", Bool(true)).Log(err.Error())
	return LoggedError{err}
}

// LogError logs the error or creates a new one using the msg if `err` is nil and returns it.
func (l *logger) LogErrorf(format string, args ...interface{}) LoggedError {
	err := fmt.Errorf(format, args...)
	return l.LogError(err)
}

type LoggedError struct {
	err error
}

func (l LoggedError) Err() error {
	return l.err
}

func (l LoggedError) Nil() error {
	return nil
}
