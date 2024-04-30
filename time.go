// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package base

import (
	"time"

	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/us"
)

const (
	// ISO8601Format represents an ISO 8601 format with timezone
	ISO8601Format = "2006-01-02T15:04:05Z07:00"
)

// Time is an time.Time struct that encodes and decodes in ISO 8601.
//
// ISO 8601 is usable by a large array of libraries whereas RFC 3339 support
// isn't often part of language standard libraries.
//
// Time also assists in calculating processing days that meet the US Federal Reserve Banks processing days.
//
// For holidays falling on Saturday, Federal Reserve Banks and Branches will be open the preceding Friday.
// For holidays falling on Sunday, all Federal Reserve Banks and Branches will be closed the following Monday.
// ACH and FedWire payments are not processed on weekends or the following US holidays.
//
// Holiday Schedule: https://www.frbservices.org/about/holiday-schedules
//
// All logic is based on ET(Eastern) time as defined by the Federal Reserve
// https://www.frbservices.org/resources/resource-centers/same-day-ach/fedach-processing-schedule.html
type Time struct {
	time.Time

	cal *cal.Calendar
}

// Now returns a Time object with the current clock time set.
func Now(location *time.Location) Time {
	// Create our calendar to attach on Time
	calendar := &cal.Calendar{
		Name: "moov-io/base",
	}
	calendar.AddHoliday(us.Holidays...) // TODO(adam): check for more?
	// calendar.Observed = cal.ObservedMonday // TODO(adam):
	return Time{
		cal:  calendar,
		Time: time.Now().In(location).Truncate(1 * time.Second),
	}
}

// NewTime wraps a time.Time value in Moov's base.Time struct.
// If you need the underlying time.Time value call .Time:
//
// The time zone will be changed to UTC.
func NewTime(t time.Time) Time {
	tt := Now(time.UTC)
	tt.Time = t // overwrite underlying Time
	return tt
}

// MarshalJSON returns JSON for the given Time
func (t Time) MarshalJSON() ([]byte, error) {
	var bs []byte
	bs = append(bs, '"')

	t.Time = t.Time.Truncate(1 * time.Second) // drop milliseconds
	bs = t.AppendFormat(bs, ISO8601Format)

	bs = append(bs, '"')
	return bs, nil
}

// UnmarshalJSON unpacks a JSON string to populate a Time instance
func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	tt, err := time.Parse(`"`+ISO8601Format+`"`, string(data))
	if err != nil || tt.IsZero() {
		// Try in RFC3339 format (default Go time)
		tt, _ = time.Parse(time.RFC3339, string(data))
		*t = NewTime(tt)
	}

	t.Time = tt.UTC().Truncate(1 * time.Second) // convert to UTC and drop millis

	return nil
}

// Equal compares two Time values. Time values are considered equal if they both truncate
// to the same year/month/day and hour/minute/second.
func (t Time) Equal(other Time) bool {
	t1 := t.Time.Truncate(1 * time.Second)
	t2 := other.Time.Truncate(1 * time.Second)
	return t1.Equal(t2)
}

func (t Time) IsHoliday() bool {
	actual, observed, _ := t.cal.IsHoliday(t.Time)

	// The Federal Reserve does not observe the following holidays on the preceding Friday
	if (!actual && observed) && t.Time.Weekday() == time.Friday {
		return false
	}

	return actual || observed
}

func (t Time) GetHoliday() *cal.Holiday {
	_, _, holiday := t.cal.IsHoliday(t.Time)
	return holiday
}

// IsBusinessDay is defined as Mondays through Fridays except federal holidays.
// Source: https://www.federalreserve.gov/Pubs/regcc/regcc.htm
func (t Time) IsBusinessDay() bool {
	actual, _, _ := t.cal.IsHoliday(t.Time)
	return !t.IsWeekend() && !actual
}

// IsBankingDay checks the rules around holidays (i.e. weekends) to determine if the given day is a banking day.
func (t Time) IsBankingDay() bool {
	// if date is not a weekend and not a holiday it is banking day.
	if t.IsWeekend() {
		return false
	}
	// and not a holiday
	if t.IsHoliday() {
		return false
	}
	// and not a monday after a holiday
	if t.Time.Weekday() == time.Monday {
		sun := t.Time.AddDate(0, 0, -1)

		actual, observed, _ := t.cal.IsHoliday(sun)
		return !actual && !observed
	}
	return true
}

// AddBusinessDay takes an integer for the number of valid business days to add and returns a Time.
// Negative values and large values (over 500 days) will not modify the Time.
func (t Time) AddBusinessDay(d int) Time {
	if d < 1 || d > 500 {
		return t
	}

	t.Time = t.Time.AddDate(0, 0, 1)
	if t.IsBusinessDay() {
		return t.AddBusinessDay(d - 1)
	}

	return t.AddBusinessDay(d)
}

// AddBankingDay takes an integer for the number of valid banking days to add and returns a Time.
// Negative values and large values (over 500 days) will not modify the Time.
func (t Time) AddBankingDay(d int) Time {
	if d < 1 || d > 500 {
		return t
	}

	t.Time = t.Time.AddDate(0, 0, 1)
	if t.IsBankingDay() {
		return t.AddBankingDay(d - 1)
	}

	return t.AddBankingDay(d)
}

// IsWeekend reports whether the given date falls on a weekend.
func (t Time) IsWeekend() bool {
	day := t.Time.Weekday()
	return day == time.Saturday || day == time.Sunday
}

// AddBankingTime increments t by the hours, minutes, and seconds provided
// but keeps the final time within 9am to 5pm in t's Location.
func (t Time) AddBankingTime(hours, minutes, seconds int) Time {
	duration := time.Duration(hours) * time.Hour
	duration += time.Duration(minutes) * time.Minute
	duration += time.Duration(seconds) * time.Second

	return addBankingDuration(t, duration)
}

func addBankingDuration(start Time, duration time.Duration) Time {
	// If we're past the current day's banking hours advance forward one day
	if start.Hour() >= 17 && (start.Minute() > 0 || start.Second() > 0) {
		start = start.AddBankingDay(1)
	}

	// Start the day at 9am or later, but not past 5pm
	if start.Hour() < 9 || start.Hour() >= 17 {
		start.Time = time.Date(start.Year(), start.Month(), start.Day(), 9, start.Minute(), start.Second(), 0, start.Location())
	}

	// Add banking hours as we can
	for duration > 0 {
		if start.IsBankingDay() {
			// Calculate the time remaining in the banking day
			endOfDay := time.Date(start.Year(), start.Month(), start.Day(), 17, 0, 0, 0, start.Location())
			remainingToday := endOfDay.Sub(start.Time)
			if duration < remainingToday {
				start.Time = start.Time.Add(duration)
				return start
			}
			duration -= remainingToday
		}
		// Move to the next banking day starting at 9 AM
		start = start.AddBankingDay(1)
		start.Time = time.Date(start.Year(), start.Month(), start.Day(), 9, 0, 0, 0, start.Location())
	}
	return start
}
