package consul

import "sync/atomic"

var globalNum int64

func AllocateUniqueID() int64 {
	atomic.AddInt64(&globalNum, 1)
	return atomic.LoadInt64(&globalNum)
}
