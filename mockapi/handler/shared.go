package handler

import (
	"sync/atomic"
)

var internalRequestCounter int64

func getNewRequestID() int64 {
	return atomic.AddInt64(&internalRequestCounter, 1)
}
