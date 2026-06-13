/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package ai

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// maxRetries is the maximum number of retry attempts after the initial request
	maxRetries = 3

	// baseBackoff is the starting delay for exponential backoff
	baseBackoff = 1 * time.Second

	// maxBackoff caps the exponential backoff delay
	maxBackoff = 30 * time.Second
)

// isRetryableStatus returns true for HTTP status codes that warrant a retry.
// This includes server errors (5xx) and rate limiting (429).
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests ||
		code == http.StatusInternalServerError ||
		code == http.StatusBadGateway ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusGatewayTimeout
}

// retryWithBackoff executes an HTTP request factory function with exponential backoff.
// The factory function should create and execute a fresh HTTP request on each call
// (since request bodies are consumed on each attempt). Returns the response on success
// or the last error if all retries are exhausted.
//
// Retry strategy:
//   - Network/timeout errors: always retried
//   - 429 Too Many Requests: retried (rate limiting)
//   - 5xx server errors: retried
//   - 4xx client errors (except 429): NOT retried (e.g., 401, 403, 404)
//   - Backoff: 1s, 2s, 4s between attempts (capped at 30s)
func retryWithBackoff(ctx context.Context, name string, factory func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait before retrying (skip on first attempt)
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseBackoff
			if delay > maxBackoff {
				delay = maxBackoff
			}
			log.Debugf("%s: retry attempt %d/%d after %v\n", name, attempt, maxRetries, delay)

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%s: context cancelled: %w", name, ctx.Err())
			case <-time.After(delay):
			}
		}

		// Execute the request
		resp, err := factory()
		if err != nil {
			lastErr = NewProviderError(name, "request failed", err)
			log.Debugf("%s: request failed (attempt %d/%d): %v\n", name, attempt, maxRetries, err)
			continue
		}

		// Check response status
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Read body for error context (body will be discarded on retry)
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Build error message
		errMsg := string(body)
		if readErr != nil {
			errMsg = fmt.Sprintf("HTTP %d (failed to read body: %v)", resp.StatusCode, readErr)
		}

		if isRetryableStatus(resp.StatusCode) {
			if resp.StatusCode == http.StatusTooManyRequests {
				log.Debugf("%s: rate limited (429), retrying...\n", name)
			} else {
				log.Debugf("%s: server error %d (attempt %d/%d), retrying...\n", name, resp.StatusCode, attempt, maxRetries)
			}
			lastErr = NewProviderError(name, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, errMsg), nil)
			continue
		}

		// Non-retryable error (4xx client errors, etc.)
		return nil, NewProviderError(name, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, errMsg), nil)
	}

	return nil, lastErr
}
