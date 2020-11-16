package log

type Fields map[string]Renderer

func (f Fields) Context() map[string]Renderer {
	return f
}
