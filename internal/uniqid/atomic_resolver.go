package uniqid

import "sync/atomic"

var lastTime int64
var lastSeq uint32

// AtomicResolver define as atomic sequence resolver, base on standard sync/atomic.
func AtomicResolver(max uint32, ms int64) (uint32, error) {
	var last int64
	var seq, localSeq uint32

	for {
		last = atomic.LoadInt64(&lastTime)
		localSeq = atomic.LoadUint32(&lastSeq)
		if last > ms {
			return max, nil
		}

		if last == ms {
			seq = max & (localSeq + 1)
			if seq == 0 {
				return max, nil
			}
		}

		if atomic.CompareAndSwapInt64(&lastTime, last, ms) && atomic.CompareAndSwapUint32(&lastSeq, localSeq, seq) {
			return (seq), nil
		}
	}
}
