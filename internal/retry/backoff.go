package retry

import (
	"math"
	"time"
)

const baseDelay = 1 * time.Second

func Backoff(attempt int32) time.Duration {
	return baseDelay * time.Duration(math.Pow(2, float64(attempt)))
}
