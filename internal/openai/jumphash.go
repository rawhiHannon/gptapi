package openai

import "sync"

func jump(key uint64, buckets int) int {
	var b int64 = -1
	var j int64 = 0
	for j < int64(buckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(1<<31) / float64(key>>33+1)))
	}
	return int(b)
}

type JumpHash struct {
	replicas int
	buckets  int
	mu       *sync.RWMutex
}

func newJumpHash(buckets, replicas int) *JumpHash {
	jhash := &JumpHash{}
	jhash.mu = new(sync.RWMutex)
	jhash.buckets = buckets
	jhash.replicas = replicas
	return jhash
}

func (j *JumpHash) get(key uint64) int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return jump(key, j.buckets)
}

func (j *JumpHash) sync(buckets int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.buckets = buckets
}
