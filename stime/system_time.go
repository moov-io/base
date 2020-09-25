package stime

import (
	"time"
)

type TimeService interface {
	Now() time.Time
}

type timeService struct{}

func NewSystemTimeService() TimeService {
	return &timeService{}
}

func (s *timeService) Now() time.Time {
	return time.Now().In(time.UTC)
}
