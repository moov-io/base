package log

type Logger interface {
	Set(key string, value Valuer) Logger
	WithPrefix(ctxs ...Context) Logger
	With(ctxs ...Context) Logger

	Debug() Logger
	Info() Logger
	Warn() Logger
	Error() Logger
	Fatal() Logger

	Log(message string)
	Logf(format string, args ...interface{})
	Send()

	LogError(error error) LoggedError
	LogErrorf(format string, args ...interface{}) LoggedError
}

type Context interface {
	Context() map[string]Valuer
}
