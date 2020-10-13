package log

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LogImplementations(t *testing.T) {
	NewDefaultLogger()
	NewNopLogger()
	NewJSONLogger()
}

func Test_Log(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Log("my message")

	a.Contains(buffer.String(), "my message")
	a.Contains(buffer.String(), "level=info")
}

func Test_WithContext(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(Error).Logf("my error message")

	a.Contains(buffer.String(), "level=error")
}

func Test_ReplaceContextValue(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(Error).With(Info).Logf("my error message")

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

	log.Set("custom", "value").Logf("test")

	a.Contains(buffer.String(), "custom=value")
}

func Test_CustomMap(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(Fields{
		"custom1": "value1",
		"custom2": "value2",
	}).Logf("test")

	output := buffer.String()
	a.Contains(output, "custom1=value1")
	a.Contains(output, "custom2=value2")
}

func Test_MultipleContexts(t *testing.T) {
	a, buffer, log := Setup(t)

	log.
		Set("custom1", "value1").
		Set("custom2", "value2").
		Logf("test")

	output := buffer.String()
	a.Contains(output, "custom1=value1")
	a.Contains(output, "custom2=value2")
}

func Test_LogError(t *testing.T) {
	a, buffer, log := Setup(t)

	newErr := errors.New("othererror")
	err := log.LogErrorf("%w", newErr)
	a.Equal("othererror", err.Error())

	output := buffer.String()
	a.Contains(output, "errored=true")
	a.Contains(output, "msg=othererror")
}

func Test_Caller(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Info().With(stacktrace).Logf("message")
	a.Regexp(regexp.MustCompile(`caller_0=(.*?)(\/log\/logger_test\.go)`), buffer.String())
}

func Setup(t *testing.T) (*assert.Assertions, *strings.Builder, Logger) {
	a := assert.New(t)
	buffer, log := NewBufferLogger()
	return a, buffer, log
}
