package log_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	lib "github.com/moov-io/base/log"
)

func Test_LogImplementations(t *testing.T) {
	lib.NewDefaultLogger()
	lib.NewNopLogger()
	lib.NewJSONLogger()
}

func Test_Log(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Log("my message")
	got := buffer.String()

	a.Contains(got, "my message")
	a.Contains(got, "level=info")

	// check valid timestamp
	tsKey := "ts="
	idx := strings.Index(got, tsKey)
	a.NotEqual(idx, -1)
	timestamp := got[idx+len(tsKey):]
	timestamp = strings.Split(timestamp, " ")[0]
	ts, err := time.Parse(time.RFC3339, timestamp)
	a.NoError(err)
	a.NotZero(ts)
}

func Test_LogWriteValue(t *testing.T) {
	uuid := uuid.New()
	barStr := "bar"
	zeroTime := time.Unix(0, 0).UTC()

	tests := []struct {
		desc     string
		key      string
		val      lib.Valuer
		expected string
	}{
		{
			key:      "foo",
			val:      lib.ByteString([]byte("bar")),
			expected: "foo=bar",
		},
		{
			key:      "foo",
			val:      lib.ByteBase64([]byte("bar")),
			expected: "foo=YmFy",
		},
		{
			key:      "foo",
			val:      lib.String(errors.New("bar").Error()),
			expected: "foo=bar",
		},
		{
			key:      "foo",
			val:      lib.Int(100),
			expected: "foo=100",
		},
		{
			key:      "foo",
			val:      lib.Int64(100),
			expected: "foo=100",
		},
		{
			key:      "foo",
			val:      lib.Int64OrNil(&number),
			expected: "foo=100",
		},
		{
			key:      "foo",
			val:      lib.Int64OrNil(nil),
			expected: "foo=null",
		},
		{
			key:      "foo",
			val:      lib.Uint32(100),
			expected: "foo=100",
		},
		{
			key:      "foo",
			val:      lib.Uint64(100),
			expected: "foo=100",
		},
		{
			key:      "foo",
			val:      lib.Float32(100),
			expected: "foo=100",
		},

		{
			key:      "foo",
			val:      lib.Float64(0.001),
			expected: "foo=0.001",
		},
		{
			key:      "foo",
			val:      lib.String("bleh"),
			expected: "foo=bleh",
		},
		{
			key:      "foo",
			val:      lib.StringOrNil(&barStr),
			expected: "foo=bar",
		},
		{
			key:      "foo",
			val:      lib.StringOrNil(nil),
			expected: "foo=null",
		},
		{
			key:      "foo",
			val:      lib.Bool(true),
			expected: "foo=true",
		},
		{
			key:      "foo",
			val:      lib.TimeDuration(time.Duration(1)),
			expected: "foo=1ns",
		},
		{
			key:      "foo",
			val:      lib.Time(zeroTime),
			expected: "foo=1970-01-01T00:00:00Z",
		},
		{
			key:      "foo",
			val:      lib.TimeOrNil(&zeroTime),
			expected: "foo=1970-01-01T00:00:00Z",
		},
		{
			key:      "foo",
			val:      lib.TimeOrNil(nil),
			expected: "foo=null",
		},
		{
			key:      "foo",
			val:      lib.TimeFormatted(time.Unix(0, 0).UTC(), time.RFC822),
			expected: "foo=\"01 Jan 70 00:00 UTC\"",
		},
		{
			key:      "foo",
			val:      lib.Stringer(uuid),
			expected: "foo=" + uuid.String(),
		},
		{
			key:      "foo",
			val:      lib.Strings([]string{"a", "b", "c"}),
			expected: "foo=\"[a, b, c]\"",
		},
	}
	for _, tc := range tests {
		a, buffer, log := Setup(t)
		log.Set(tc.key, tc.val).Send()
		got := buffer.String()
		a.Contains(got, tc.expected)
	}
}

func Test_Send(t *testing.T) {
	a, buffer, log := Setup(t)
	log.Set("foo", lib.String("bar")).Send()

	got := buffer.String()
	a.NotContains(got, "msg=")
	a.Contains(got, "ts=")
	a.Contains(got, "foo=bar")
}

func Test_WithContext(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(lib.Error).Logf("my error message")

	a.Contains(buffer.String(), "level=error")
}

func Test_ReplaceContextValue(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(lib.Error).Warn().Logf("my error message")

	a.Contains(buffer.String(), "level=warn")
}

func Test_Debug(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Debug().Logf("message")

	a.Contains(buffer.String(), "level=debug")
}

func Test_Info(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Info().Logf("message")

	a.Contains(buffer.String(), "level=info")
}

func Test_Error(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Error().Logf("message")

	a.Contains(buffer.String(), "level=error")
}

func Test_ErrorF(t *testing.T) {
	a, buffer, log := Setup(t)

	err := errors.New("error")
	log.Error().LogErrorf("message %w", err)

	a.Contains(buffer.String(), "msg=\"message error\"")
	a.Contains(buffer.String(), "errored=true")
}

func Test_Fatal(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Fatal().Logf("message")

	a.Contains(buffer.String(), "level=fatal")
}

func Test_CustomKeyValue(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Set("custom", lib.String("value")).Logf("test")

	a.Contains(buffer.String(), "custom=value")
}

func Test_CustomMap(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(lib.Fields{
		"custom1": lib.String("value1"),
		"custom2": lib.String("value2"),
	}).Logf("test")

	output := buffer.String()
	a.Contains(output, "custom1=value1")
	a.Contains(output, "custom2=value2")
}

func Test_MultipleContexts(t *testing.T) {
	a, buffer, log := Setup(t)

	log.
		Set("custom1", lib.String("value1")).
		Set("custom2", lib.String("value2")).
		Logf("test")

	output := buffer.String()
	a.Contains(output, "custom1=value1")
	a.Contains(output, "custom2=value2")
}

func Test_LogError(t *testing.T) {
	a, buffer, log := Setup(t)

	newErr := errors.New("othererror")
	err := log.LogErrorf("wrap: %w", newErr).Err()
	a.Equal("wrap: othererror", err.Error())

	output := buffer.String()
	a.Contains(output, "errored=true")
	a.Contains(output, "msg=\"wrap: othererror\"")

	wrappedErr := fmt.Errorf("wrapped: %w", newErr)
	gotErr := log.LogError(wrappedErr).Err()
	a.True(errors.Is(gotErr, newErr))
}

func Test_Caller(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Info().With(lib.StackTrace).Logf("message")
	a.Regexp(regexp.MustCompile(`caller_0=(.*?)(\/log\/logger_test\.go)`), buffer.String())
}

func Setup(t *testing.T) (*assert.Assertions, *strings.Builder, lib.Logger) {
	a := assert.New(t)
	buffer, log := lib.NewBufferLogger()
	return a, buffer, log
}
