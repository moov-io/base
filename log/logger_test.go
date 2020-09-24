package log

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LogImplementations(t *testing.T) {
	NewDefaultLogger()
	NewNopLogger()
}

func Test_Log(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Log("my message")

	a.Contains(buffer.String(), "my message")
	a.Contains(buffer.String(), "level=info")
}

func Test_WithContext(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(Error).Log("my error message")

	a.Contains(buffer.String(), "level=error")
}

func Test_ReplaceContextValue(t *testing.T) {
	a, buffer, log := Setup(t)

	log.With(Error).With(Info).Log("my error message")

	a.Contains(buffer.String(), "level=info")
}

func Test_Info(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Info().Log("message")

	a.Contains(buffer.String(), "level=info")
}

func Test_Error(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Error().Log("message")

	a.Contains(buffer.String(), "level=error")
}

func Test_ErrorF(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Error().LogErrorF("message %w", errors.New("error"))

	a.Contains(buffer.String(), "msg=\"message error\"")
	a.Contains(buffer.String(), "error=\"message error\"")
}

func Test_Fatal(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Fatal().Log("message")

	a.Contains(buffer.String(), "level=fatal")
}

func Test_CustomKeyValue(t *testing.T) {
	a, buffer, log := Setup(t)

	log.WithKeyValue("custom", "value").Log("test")

	a.Contains(buffer.String(), "custom=value")
}

func Test_CustomMap(t *testing.T) {
	a, buffer, log := Setup(t)

	log.WithMap(map[string]string{
		"custom1": "value1",
		"custom2": "value2",
	}).Log("test")

	output := buffer.String()
	a.Contains(output, "custom1=value1")
	a.Contains(output, "custom2=value2")
}

func Test_MultipleContexts(t *testing.T) {
	a, buffer, log := Setup(t)

	log.
		WithKeyValue("custom1", "value1").
		WithKeyValue("custom2", "value2").
		Log("test")

	output := buffer.String()
	a.Contains(output, "custom1=value1")
	a.Contains(output, "custom2=value2")
}

func Test_LogErrorNil(t *testing.T) {
	a, buffer, log := Setup(t)

	err := log.LogError("someerror", nil)
	a.Equal("someerror", err.Error())

	output := buffer.String()
	a.Contains(output, "error=someerror")
	a.Contains(output, "msg=someerror")
}

func Test_LogError(t *testing.T) {
	a, buffer, log := Setup(t)

	err := log.LogError("someerror", errors.New("othererror"))
	a.Equal("othererror", err.Error())

	output := buffer.String()
	a.Contains(output, "error=othererror")
	a.Contains(output, "msg=someerror")
}

func Test_Caller(t *testing.T) {
	a, buffer, log := Setup(t)

	log.Info().Log("message")

	a.Contains(buffer.String(), "caller_0=")

}

func Setup(t *testing.T) (*assert.Assertions, *strings.Builder, Logger) {
	a := assert.New(t)
	buffer, log := NewBufferLogger()
	return a, buffer, log
}
