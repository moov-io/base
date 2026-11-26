package base_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/moov-io/base"
)

func ExampleID() {
	id := base.ID()
	fmt.Println(len(id) == 40)
	// Output:
	// true
}

func ExampleErrorList() {
	first := errors.New("first")
	var list base.ErrorList
	list.Add(first)
	list.Add(errors.New("second"))

	// Has and Match compare simple errors by identity, not only by message.
	fmt.Println(base.Has(list, first))
	fmt.Println(list.Err())
	// Output:
	// true
	// first
}

func ExampleTime() {
	eastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	// Thanksgiving Day 2026 is a Federal Reserve holiday.
	day := base.NewTime(time.Date(2026, time.November, 26, 10, 0, 0, 0, eastern))
	fmt.Println(day.IsHoliday(), day.IsBankingDay())
	fmt.Println(day.AddBankingDay(1).Format("2006-01-02"))

	// Independence Day 2026 is Saturday. The Fed is open on Friday 2026-07-03.
	friday := base.NewTime(time.Date(2026, time.July, 3, 10, 0, 0, 0, eastern))
	holiday := friday.GetHoliday()
	fmt.Println(holiday.Name, friday.IsHoliday(), friday.IsBankingDay())
	// Output:
	// true false
	// 2026-11-27
	// Independence Day false true
}
