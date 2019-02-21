package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {

	t.Run("Test Memory Cache", func(t *testing.T) {
		cache := New()

		duration, _ := time.ParseDuration("1h")
		cache.Set("key", "value", duration)
		value, ok := cache.Get("key")

		if !ok {
			t.Errorf("Failed to retrieve cached value")
		}

		if value != "value" {
			t.Errorf("Unexpected value returned from cached data: %s", value)
		}

		duration, _ = time.ParseDuration("-1s")
		cache.Set("key", "value", duration)
		value, ok = cache.Get("key")

		if ok {
			t.Errorf("Cache failed to expire")
		}
	})
}
