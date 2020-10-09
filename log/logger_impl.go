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
		writer:  writer,
		ctx:     map[string]string{},
		lastErr: nil,
	}

	// Default logs to be info until changed
	return l.Info()
}

type logger struct {
	writer  log.Logger
	ctx     map[string]string
	lastErr error
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
		writer:  l.writer,
		ctx:     combined,
		lastErr: l.lastErr,
	}
}

func (l *logger) WithError(err error) Logger {
	ctx := &logger{
		writer:  l.writer,
		ctx:     l.ctx,
		lastErr: err,
	}

	if err != nil {
		return ctx.Set("error", err.Error())
	} else {
		return ctx.Set("error", "")
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

func (l *logger) Log(format string, a ...interface{}) {
	keyvals := make([]interface{}, (len(l.ctx)*2)+2)

	keyvals[0] = "msg"
	keyvals[1] = fmt.Sprintf(format, a...)

	i := 2
	for k, v := range l.ctx {
		keyvals[i] = k
		keyvals[i+1] = v
		i += 2
	}

	l.writer.Log(keyvals...)
}

// LogError logs the error or creates a new one using the msg if `err` is nil and returns it.
func (l *logger) LogError(format string, a ...interface{}) error {
	newErr := fmt.Errorf(format, a...)

	if l.lastErr == nil {
		return l.WithError(newErr).LogError(format, a...)
	} else {
		l.Log(newErr.Error())
		return l.lastErr
	}
}
