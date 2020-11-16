package log

import (
	"fmt"
	"runtime"
	"strings"
)

type st string

const StackTrace = st("stacktrace")

// Context returns the map that states that key value of `level={{l}}`
func (s st) Context() map[string]Valuer {
	kv := map[string]Valuer{}

	i := 0
	c := 0
	_, file, line, ok := runtime.Caller(i)
	for ; ok; i++ {
		if c > 0 || (!strings.HasSuffix(file, "model_stacktrace.go") && !strings.HasSuffix(file, "logger_impl.go")) {
			key := fmt.Sprintf("caller_%d", c)
			value := fmt.Sprintf("%s:%d", file, line)
			kv[key] = String(value)
			c++
		}
		_, file, line, ok = runtime.Caller(i + 1)
	}

	return kv
}
