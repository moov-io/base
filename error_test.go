package base

import (
	"bytes"
	"errors"
	"testing"
)

func TestErrorAdd(t *testing.T) {

	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))

	es := errorList.Error()

	if es != "testing" {
		t.Errorf("got %s", errorList.Error())
	}

	if errorList.Empty() {
		t.Errorf("ErrorList is empty: %v", errorList)
	}

	errorList.Add(errors.New("continued testing"))

	if errorList.Empty() {
		t.Errorf("ErrorList is empty: %v", errorList)
	}
}

func TestErrorErr(t *testing.T) {
	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))

	e1 := errorList.Err()

	if e1.Error() != "testing" {
		t.Errorf("got %q", e1)
	}

}

func TestErrorPrint(t *testing.T) {
	var buf bytes.Buffer

	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))
	errorList.Add(errors.New("continued testing"))

	// nil
	errorList.Print(&buf)
	if v := buf.String(); v == "<nil>" {
		t.Errorf("got %q", v)
	}
	if v := errorList.Error(); v == "<nil>" {
		t.Errorf("got %q", v)
	}
	buf.Reset()

}

func TestErrorEmpty(t *testing.T) {
	errorList := ErrorList{}

	e1 := errorList.Err()

	if e1 != nil {
		t.Errorf("got %q", e1)
	}

	if errorList.Error() != "<nil>" {
		t.Errorf("got %s", errorList.Error())
	}

	var buf bytes.Buffer

	errorList.Print(&buf)
}
