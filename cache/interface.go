package cache

import (
	"time"
)

// Simple cache interface with Get and Set methods.
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}
