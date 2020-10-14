package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
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

func NewBufferLogger() (*strings.Builder, Logger) {
	buffer := strings.Builder{}
	writer := log.NewLogfmtLogger(log.NewSyncWriter(&buffer))
	log := NewLogger(writer)
	return &buffer, log
}

func NewLogger(writer log.Logger) Logger {
	l := &logger{
		writer: writer,
		ctx:    map[string]string{},
	}

	// Default logs to be info until changed
	return l.Info()
}

type logger struct {
	writer log.Logger
	ctx    map[string]string
}

func (l *logger) Set(key, value string) Logger {
	return l.With(Fields{
		key: value,
	})
}

// With returns a new Logger with the contexts added to its own.
func (l *logger) With(ctxs ...Context) Logger {
	// Estimation assuming that for each ctxs has at least 1 value.
	combined := make(map[string]string, len(l.ctx)+len(ctxs))

	for k, v := range l.ctx {
		combined[k] = v
	}

	for _, c := range ctxs {
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
	l.Logf(msg)
}

func (l *logger) Logf(format string, args ...interface{}) {
	keyvals := make([]interface{}, (len(l.ctx)*2)+2)

	keyvals[0] = "msg"
	keyvals[1] = fmt.Sprintf(format, args...)

	i := 2
	for k, v := range l.ctx {
		keyvals[i] = k
		keyvals[i+1] = v
		i += 2
	}

	_ = l.writer.Log(keyvals...)
}

func (l *logger) LogError(err error) error {
	return l.LogErrorf(err.Error())
}

// LogError logs the error or creates a new one using the msg if `err` is nil and returns it.
func (l *logger) LogErrorf(format string, args ...interface{}) error {
	newErr := fmt.Errorf(format, args...)
	l.Set("errored", "true").Logf(newErr.Error())
	return newErr
}
