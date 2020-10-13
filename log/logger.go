package log

type Logger interface {
	Set(key, value string) Logger
	With(ctxs ...Context) Logger

	Info() Logger
	Warn() Logger
	Error() Logger
	Fatal() Logger

	Logf(format string, a ...interface{})
	LogErrorf(format string, a ...interface{}) error
}
