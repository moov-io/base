// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package base

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))

	pse := ParseError{
		Err:    errorList,
		Line:   5,
		Record: "ABC",
	}

	if !strings.Contains(pse.Error(), "testing") {
		t.Errorf("got %s", errorList.Error())
	}

	if pse.Record != "ABC" {
		t.Errorf("got %s", pse.Record)
	}

	if pse.Line != 5 {
		t.Errorf("got %v", pse.Line)
	}

}

func TestParseErrorRecordNull_Error(t *testing.T) {
	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))

	pse := ParseError{
		Err:    errorList,
		Line:   5,
		Record: "",
	}

	e1 := pse.Error()

	if e1 != "line:5 base.ErrorList testing" {
		t.Errorf("got %s", e1)
	}
}

func TestErrorList_Add(t *testing.T) {
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

func TestErrorList_Err(t *testing.T) {
	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))

	e1 := errorList.Err()

	if e1.Error() != "testing" {
		t.Errorf("got %q", e1)
	}

}

func TestErrorList_Print(t *testing.T) {
	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))
	errorList.Add(errors.New("continued testing"))

	var buf bytes.Buffer
	errorList.Print(&buf)

	if v := errorList.Error(); v == "<nil>" {
		t.Errorf("got %q", v)
	}
	buf.Reset()

}

func TestErrorList_Empty(t *testing.T) {
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
	buf.Reset()
}

func TestErrorList_MarshalJSON(t *testing.T) {
	errorList := ErrorList{}
	errorList.Add(errors.New("testing"))
	errorList.Add(errors.New("continued testing"))
	errorList.Add(errors.New("testing again"))
	errorList.Add(errors.New("continued testing again"))

	b, err := errorList.MarshalJSON()

	if len(b) == 0 {
		t.Errorf("got %s", errorList.Error())
	}
	if err != nil {
		t.Errorf("got %s", errorList.Error())
	}
}
