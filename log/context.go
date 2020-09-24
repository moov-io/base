package log

type Context interface {
	Context() map[string]string
}
