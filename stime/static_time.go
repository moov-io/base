package stime

import (
	"time"
)

type StaticTimeService interface {
	Change(update time.Time) time.Time
	Add(d time.Duration) time.Time
	TimeService
}

type staticTimeService struct {
	time time.Time
}

func NewStaticTimeService() StaticTimeService {
	return &staticTimeService{
		time: time.Now().In(time.UTC).Round(time.Second),
	}
}

func (s *staticTimeService) Now() time.Time {
	return s.time
}

func (s *staticTimeService) Change(update time.Time) time.Time {
	s.time = update
	return s.time
}

func (s *staticTimeService) Add(d time.Duration) time.Time {
	s.time = s.time.Add(d)
	return s.time
}
