package log

type Context interface {
	Context() map[string]string
}

type Fields map[string]string

var _ Context = Fields{}

func (f Fields) Context() map[string]string {
	return f
}
