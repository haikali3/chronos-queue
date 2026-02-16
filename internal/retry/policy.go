package retry

import "time"

func ShouldRetry(retryCount, maxRetries int32) bool {
	return retryCount < maxRetries
}

func NextRetryAt(attempt int32) time.Time {
	return time.Now().Add(Backoff(attempt))
}
