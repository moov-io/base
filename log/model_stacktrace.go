package log

import (
	"fmt"
	"runtime"
	"strings"
)

type st string

// Fatal sets level=fatal in the log output
const StackTrace = st("stacktrace")

// Context returns the map that states that key value of `level={{l}}`
func (s st) Context() map[string]string {
	kv := map[string]string{}

	i := 0
	c := 0
	_, file, line, ok := runtime.Caller(i)
	for ; ok; i++ {
		if c > 0 || !strings.HasSuffix(file, "logger.go") {
			key := fmt.Sprintf("caller_%d", c)
			value := fmt.Sprintf("%s:%d", file, line)
			kv[key] = value
			c++
		}
		_, file, line, ok = runtime.Caller(i + 1)
	}

	return kv
}
