package log

import (
	"encoding/base64"
	"time"
)

// Renderer is an interface to deal with typing problems of just having an interface{} as the acceptable parameters
// Go-kit logging has a failure case if you attempt to throw any values into it.
// This is a way to guard our developers from having to worry about error cases of the lower logging framework.
type Renderer interface {
	getValue() interface{}
}

type any struct {
	value interface{}
}

func (a *any) getValue() interface{} {
	return a.value
}

func String(s string) Renderer {
	return &any{s}
}

func Int(i int) Renderer {
	return &any{i}
}

func Float64(f float64) Renderer {
	return &any{f}
}

func Bool(b bool) Renderer {
	return &any{b}
}

func TimeDuration(d time.Duration) Renderer {
	return &any{d.String()}
}

func Time(t time.Time) Renderer {
	return TimeFormatted(t, time.RFC3339Nano)
}

func TimeFormatted(t time.Time, format string) Renderer {
	return String(t.Format(format))
}

func ByteString(b []byte) Renderer {
	return String(string(b))
}

func ByteBase64(b []byte) Renderer {
	return String(base64.RawURLEncoding.EncodeToString(b))
}
