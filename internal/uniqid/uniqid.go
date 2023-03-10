package uniqid

import (
	"time"
)

const (
	TimestampLength = 39
	SequenceLength  = 24
	MaxSequence     = 1<<SequenceLength - 1
	MaxTimestamp    = 1<<TimestampLength - 1

	timestampMoveLength = SequenceLength
)

type SequenceResolver func(max uint32, ms int64) (uint32, error)

var (
	resolver  SequenceResolver
	startTime = time.Date(2004, 11, 10, 23, 0, 0, 0, time.UTC)
)

func NextId() uint64 {
	c := currentMillis()
	seqResolver := callSequenceResolver()
	seq, err := seqResolver(MaxSequence, c)
	if err != nil {
		return 0
	}
	for seq >= MaxSequence {
		c = waitForNextMillis(c)
		seq, err = seqResolver(MaxSequence, c)
		if err != nil {
			return 0
		}
	}
	//TODO: Think about the life time
	df := int(elapsedTime(c, startTime))
	id := uint64((df << timestampMoveLength) | int(seq))
	return id
}

func waitForNextMillis(last int64) int64 {
	now := currentMillis()
	for now == last {
		now = currentMillis()
	}
	return now
}

func callSequenceResolver() SequenceResolver {
	if resolver == nil {
		return AtomicResolver
	}

	return resolver
}

func elapsedTime(nowms int64, s time.Time) int64 {
	return nowms - s.UTC().UnixNano()/1e6
}

// currentMillis get current millisecond.
func currentMillis() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}
