package log

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// Valuer is an interface to deal with typing problems of just having an interface{} as the acceptable parameters
// Go-kit logging has a failure case if you attempt to throw any values into it.
// This is a way to guard our developers from having to worry about error cases of the lower logging framework.
type Valuer interface {
	getValue() interface{}
}

type any struct {
	value interface{}
}

func (a *any) getValue() interface{} {
	return a.value
}

func String(s string) Valuer {
	return &any{s}
}

func StringOrNil(s *string) Valuer {
	if s == nil {
		return &any{nil}
	}
	return String(*s)
}

func Int(i int) Valuer {
	return &any{i}
}

func Int64(i int64) Valuer {
	return &any{i}
}

func Int64OrNil(i *int64) Valuer {
	return &any{i}
}

func Uint32(i uint32) Valuer {
	return &any{i}
}

func Uint64(i uint64) Valuer {
	return &any{i}
}

func Float32(f float32) Valuer {
	return &any{f}
}

func Float64(f float64) Valuer {
	return &any{f}
}

func Bool(b bool) Valuer {
	return &any{b}
}

func TimeDuration(d time.Duration) Valuer {
	return &any{d.String()}
}

func Time(t time.Time) Valuer {
	return TimeFormatted(t, time.RFC3339Nano)
}

func TimeOrNil(t *time.Time) Valuer {
	if t == nil {
		return &any{nil}
	}
	return Time(*t)
}

func TimeFormatted(t time.Time, format string) Valuer {
	return String(t.Format(format))
}

func ByteString(b []byte) Valuer {
	return String(string(b))
}

func ByteBase64(b []byte) Valuer {
	return String(base64.RawURLEncoding.EncodeToString(b))
}

func Stringer(s fmt.Stringer) Valuer {
	return &any{s.String()}
}

func Strings(vals []string) Valuer {
	out := fmt.Sprintf("[%s]", strings.Join(vals, ", "))
	return String(out)
}
