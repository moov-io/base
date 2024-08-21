// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package base

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestErrorList__EmptyThenNot(t *testing.T) {
	var el ErrorList
	require.NoError(t, el.Err())
	require.Equal(t, "<nil>", el.Error())
	require.True(t, el.Empty())

	el.Add(errors.New("bad thing"))
	require.Error(t, el.Err())
	require.Equal(t, "bad thing", el.Error())
	require.False(t, el.Empty())
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

// testMatch validates the Match error function
func TestMatch(t *testing.T) {
	testError := errors.New("Test error")

	if !Match(nil, nil) {
		t.Error("Match should be reflexive on nil")
	}

	if !Match(testError, testError) {
		t.Error("Match should be reflexive")
	}

	p := ParseError{Err: testError}
	if !Match(p, testError) {
		t.Error("Match should match wrapped errors implementing the UnwrappableError interface")
	}

	differentError := errors.New("Different error")
	if Match(testError, differentError) {
		t.Error("Match should return false for different simple errors")
	}

	q := ParseError{Err: differentError}
	if !Match(p, q) {
		t.Error("Match should match two different ParseErrors to each other since they have the same type")
	}

	errorList := ErrorList{}
	if Match(errorList, p) {
		t.Error("Match should return false for errors with different types")
	}
}

// testHas validates the Has error function
func TestHas(t *testing.T) {
	err := errors.New("Non list error")

	if Has(err, err) {
		t.Error("Has should return false when given a non-list error as the first arg")
	}

	if Has(nil, err) {
		t.Error("Has should not return true if there are no errors")
	}

	if Has(ErrorList([]error{}), err) {
		t.Error("Has should not return true if there are no errors")
	}

	if !Has(ErrorList([]error{err}), err) {
		t.Error("Has should return true if the error list has the test error")
	}
}

func TestErrorList_Panic(t *testing.T) {
	var el ErrorList
	require.Equal(t, "<nil>", fmt.Sprintf("%v", el))
	require.Equal(t, "<nil>", fmt.Errorf("%w", el).Error())
}
