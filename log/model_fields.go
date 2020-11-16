package log

type Fields map[string]Valuer

func (f Fields) Context() map[string]Valuer {
	return f
}
