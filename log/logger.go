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

	LogError(error error) LoggedError
	LogErrorf(format string, args ...interface{}) LoggedError
}

type Context interface {
	Context() map[string]string
}
