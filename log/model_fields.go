package log

type Fields map[string]string

func (f Fields) Context() map[string]string {
	return f
}
