package repository

import (
	mockableClock "github.com/stephanos/clock"
	"time"
)

var clock = mockableClock.New()

func makeTimestamp() int64 {
	return clock.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
