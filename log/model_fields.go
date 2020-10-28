package log

type Fields map[string]interface{}

func (f Fields) Context() map[string]interface{} {
	return f
}
