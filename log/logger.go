package log

type Logger interface {
	Set(key, value string) Logger
	With(ctxs ...Context) Logger

	Info() Logger
	Warn() Logger
	Error() Logger
	Fatal() Logger

	Log(message string)
	Logf(format string, args ...interface{})

	LogError(error error)
	LogErrorf(format string, args ...interface{}) error
}

type Context interface {
	Context() map[string]string
}
