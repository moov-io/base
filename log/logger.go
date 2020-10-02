package log

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/go-kit/kit/log"
)

type Logger interface {
	With(ctxs ...Context) Logger
	WithMap(mapCtx map[string]string) Logger
	WithKeyValue(keyvals ...string) Logger

	Info() Logger
	Error() Logger
	Fatal() Logger

	Log(msg string)
	LogError(msg string, err error) error
	LogErrorF(format string, a ...interface{}) error
}

type logger struct {
	writer log.Logger
	ctx    map[string]string
}

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

func (l *logger) WithMap(mapCtx map[string]string) Logger {
	// Estimation assuming that for each ctxs has at least 1 value.
	combined := make(map[string]string, len(l.ctx)+len(mapCtx))

	for k, v := range l.ctx {
		combined[k] = v
	}

	for k, v := range mapCtx {
		combined[k] = v
	}

	return &logger{
		writer: l.writer,
		ctx:    combined,
	}
}

func (l *logger) WithKeyValue(keyvals ...string) Logger {
	input := map[string]string{}

	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, log.ErrMissingValue.Error())
	}

	// split into key/value pars
	for i := 0; i < len(keyvals); i += 2 {
		input[keyvals[i]] = keyvals[i+1]
	}

	return l.WithMap(input)
}

func (l *logger) Info() Logger {
	return l.With(Info)
}

func (l *logger) Error() Logger {
	return l.With(Error)
}

func (l *logger) Fatal() Logger {
	return l.With(Fatal)
}

func (l *logger) Log(msg string) {
	i := 0
	keyvals := make([]interface{}, (len(l.ctx)*2)+2)
	for k, v := range l.ctx {
		keyvals[i] = k
		keyvals[i+1] = v
		i += 2
	}

	keyvals[i] = "msg"
	keyvals[i+1] = msg

	i = 0
	c := 0
	_, file, line, ok := runtime.Caller(i)
	for ; ok; i++ {
		if c > 0 || !strings.HasSuffix(file, "logger.go") {
			keyvals = append(keyvals, fmt.Sprintf("caller_%d", c), fmt.Sprintf("%s:%d", file, line))
			c++
		}
		_, file, line, ok = runtime.Caller(i + 1)
	}

	l.writer.Log(keyvals...)
}

// LogError logs the error or creates a new one using the msg if `err` is nil and returns it.
func (l *logger) LogError(msg string, err error) error {
	if err == nil {
		err = errors.New(msg)
	}

	l.WithKeyValue("error", err.Error()).Log(msg)
	return err
}

func (l *logger) LogErrorF(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	return l.LogError(err.Error(), err)
}
