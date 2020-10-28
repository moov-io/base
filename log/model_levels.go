package log

// Level just wraps a string to be able to add Context specific to log levels
type Level string

// Info is sets level=info in the log output
const Info = Level("info")

const Warn = Level("warn")

// Error sets level=error in the log output
const Error = Level("error")

// Fatal sets level=fatal in the log output
const Fatal = Level("fatal")

// Context returns the map that states that key value of `level={{l}}`
func (l Level) Context() map[string]interface{} {
	return map[string]interface{}{
		"level": string(l),
	}
}
