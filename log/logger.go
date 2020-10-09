package log

type Logger interface {
	Set(key, value string) Logger
	With(ctxs ...Context) Logger

	WithError(err error) Logger

	Info() Logger
	Warn() Logger
	Error() Logger
	Fatal() Logger

	Log(format string, a ...interface{})
	LogError(format string, a ...interface{}) error
}
