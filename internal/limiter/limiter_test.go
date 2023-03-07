package limiter

import (
	"gptapi/internal/idgen"
	"testing"
	"time"
)

func TestRedisRateLimiter(t *testing.T) {
	limiter := NewRedisRateLimiter("test_", 20, 1*time.Minute)
	key := idgen.NextId()
	for i := 1; i <= 20; i++ {
		if !limiter.Allow(key) {
			t.Errorf("allow request failed at attempt %d", i)
		}
	}
	if limiter.Allow(key) {
		t.Error("allow request succeeded after reaching the limit")
	}
	time.Sleep(1 * time.Minute)
	if !limiter.Allow(key) {
		t.Error("allow request failed after the duration has expired")
	}
}
