package log_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	lib "github.com/moov-io/base/log"
	"github.com/stretchr/testify/assert"
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
	tests := []struct {
		desc     string
		key      string
		val      lib.Renderer
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
			val:      lib.String("bleh"),
			expected: "foo=bleh",
		},
		{
			key:      "foo",
			val:      lib.Float64(0.001),
			expected: "foo=0.001",
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
			val:      lib.Time(time.Unix(0, 0)),
			expected: "foo=1969-12-31T18:00:00-06:00",
		},
		{
			key:      "foo",
			val:      lib.TimeFormatted(time.Unix(0, 0), time.RFC822),
			expected: "foo=\"31 Dec 69 18:00 CST\"",
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

	log.With(lib.Error).With(lib.Info).Logf("my error message")

	a.Contains(buffer.String(), "level=info")
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
