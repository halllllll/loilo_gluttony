package scrape

import "github.com/cenkalti/backoff"

func withRetry(operation func() error) error {
	b := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)
	return backoff.Retry(operation, b)
}
