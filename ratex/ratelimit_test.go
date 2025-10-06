package ratex

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestExecRetryable(t *testing.T) {
	ctx := context.Background()

	t.Run("Success on first try", func(t *testing.T) {
		closure := func(ctx context.Context) (string, error) {
			return "success", nil
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.NoErrorf(t, err, "Expected success, got error: %v", err)
		require.Equalf(t, "success", result, "Expected result 'success', got: %v", result)
	})

	t.Run("Retryable failure with success before last retry", func(t *testing.T) {
		attempts := 0
		closure := func(ctx context.Context) (string, error) {
			attempts++
			if attempts < 3 {
				return "", errors.New("retryable error")
			}
			return "success", nil
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		start := time.Now()
		result, err := ExecRetryable(ctx, closure, params)
		elapsed := time.Since(start)
		require.NoErrorf(t, err, "Expected success, got error: %v", err)
		require.Equalf(t, "success", result, "Expected result 'success', got: %v", result)
		require.Equal(t, 3, attempts, "Expected 3 attempts")
		// Check approximate backoff time (2 backoffs: ~10-50ms *1 + ~10-50ms *2)
		minElapsed := params.MinDuration*1 + params.MinDuration*2
		maxElapsed := params.MaxDuration*1 + params.MaxDuration*2 + 50*time.Millisecond // Overhead allowance
		require.GreaterOrEqual(t, elapsed, minElapsed, "Elapsed time too short")
		require.LessOrEqual(t, elapsed, maxElapsed, "Elapsed time too long")
	})

	t.Run("Non-retryable failure", func(t *testing.T) {
		closure := func(ctx context.Context) (string, error) {
			return "", errors.New("non-retryable error")
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return false },
			MaxAttempts: 3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.Errorf(t, err, "Expected non-retryable error, got: %v", err)
		require.Empty(t, result)
		require.Equal(t, "non-retryable error", err.Error())
	})

	t.Run("Retryable failures exceeding MaxAttempts", func(t *testing.T) {
		attempts := 0
		closure := func(ctx context.Context) (string, error) {
			attempts++
			return "", errors.New("retryable error")
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.Errorf(t, err, "Expected error after exceeding max retries, got: %v", err)
		require.Empty(t, result)
		require.Equal(t, 3, attempts, "Expected 3 attempts")
		require.Equal(t, "hit max tries 3: try 3 of 3: retryable error", err.Error())
	})

	t.Run("Context cancellation during closure", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		attempts := 0
		closure := func(ctx context.Context) (string, error) {
			attempts++
			if attempts == 2 {
				cancel()
			}
			return "", errors.New("retryable error")
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.ErrorIs(t, err, context.Canceled)
		require.Empty(t, result)
		require.Equal(t, 2, attempts, "Expected 2 attempts before cancel")
	})

	t.Run("Context cancellation during backoff", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		closure := func(ctx context.Context) (string, error) {
			return "", errors.New("retryable error")
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 3,
			MinDuration: 100 * time.Millisecond,
			MaxDuration: 200 * time.Millisecond,
		}
		go func() {
			time.Sleep(50 * time.Millisecond) // Cancel mid-backoff
			cancel()
		}()
		result, err := ExecRetryable(ctx, closure, params)
		require.ErrorIs(t, err, context.Canceled)
		require.Empty(t, result)
	})

	t.Run("Invalid params with defaults", func(t *testing.T) {
		closure := func(ctx context.Context) (string, error) {
			return "success", nil
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 0, // Should default to 1
			MinDuration: 0, // Should default to 100ms
			MaxDuration: 0, // Should default to 100ms * 10 = 1s
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.NoError(t, err)
		require.Equal(t, "success", result)
		// Note: Defaults are applied, but no retries needed here
	})

	t.Run("MinDuration > MaxDuration", func(t *testing.T) {
		closure := func(ctx context.Context) (string, error) {
			return "success", nil
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 3,
			MinDuration: 100 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond, // Will be set to 100ms * 10 = 1s
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.NoError(t, err)
		require.Equal(t, "success", result)
	})

	t.Run("Different return type", func(t *testing.T) {
		closure := func(ctx context.Context) (int, error) {
			return 42, nil
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxAttempts: 3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.NoError(t, err)
		require.Equal(t, 42, result)
	})
}

func TestGenerateRateLimitDuration(t *testing.T) {
	t.Run("Standard case", func(t *testing.T) {
		for i := 0; i < 10; i++ { // Run multiple times to check randomness
			dur, err := generateRateLimitDuration(1, 100*time.Millisecond, 200*time.Millisecond)
			require.NoError(t, err)
			require.GreaterOrEqual(t, dur, 100*time.Millisecond)
			require.LessOrEqual(t, dur, 200*time.Millisecond)
		}
	})

	t.Run("Max <= Min", func(t *testing.T) {
		dur, err := generateRateLimitDuration(2, 100*time.Millisecond, 50*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, 200*time.Millisecond, dur) // min * multiplier
	})

	t.Run("Delta == 0", func(t *testing.T) {
		dur, err := generateRateLimitDuration(3, 50*time.Millisecond, 50*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, 150*time.Millisecond, dur)
	})

	t.Run("Overflow cap", func(t *testing.T) {
		// Set large multiplier to trigger cap
		// Assume cap at math.MaxInt64 / 1000000 ~ 9e12 ms (~104 days)
		// Use minVal=1ms, multiplier such that 1 * mul > 9e12
		largeMul := int(math.MaxInt64 / 1000000 / 2) // Safe large int
		dur, err := generateRateLimitDuration(largeMul, 1*time.Millisecond, 2*time.Millisecond)
		require.NoError(t, err)
		// Since min + rand(0 or 1) * largeMul, but cap to max * mul = 2 * largeMul ms
		require.LessOrEqual(t, int64(dur.Milliseconds()), 2*int64(largeMul))
	})

	t.Run("Negative multiplier (edge case)", func(t *testing.T) {
		dur, err := generateRateLimitDuration(-1, 100*time.Millisecond, 200*time.Millisecond)
		require.NoError(t, err)
		// Since waitInterval negative, cap kicks in, but code sets to max * mul if overflow/negative
		// But mul negative, so waitInterval negative, capped to max * mul (negative, but duration cast)
		// Actually, code checks waitInterval <0, sets to maxVal * int64(multiplier)
		// If mul negative, this would be negative, but time.Duration negative is invalid
		// Note: This test highlights potential issue, but multiplier is always positive in usage
		require.LessOrEqual(t, dur, time.Duration(0)) // Expect non-positive
	})
}

func TestRateLimit(t *testing.T) {
	ctx := context.Background()

	t.Run("New limiter", func(t *testing.T) {
		params := RateLimitParams{
			RateLimiter: nil,
			Attempt:     1,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 20 * time.Millisecond,
		}
		limiter, err := RateLimit(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, limiter)
	})

	t.Run("Existing limiter update", func(t *testing.T) {
		existing := rate.NewLimiter(rate.Every(100*time.Millisecond), 1)
		params := RateLimitParams{
			RateLimiter: existing,
			Attempt:     2,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 20 * time.Millisecond,
		}
		limiter, err := RateLimit(ctx, params)
		require.NoError(t, err)
		require.Equal(t, existing, limiter)
	})

	t.Run("Context cancel during wait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		params := RateLimitParams{
			RateLimiter: nil,
			Attempt:     1,
			MinDuration: 100 * time.Millisecond,
			MaxDuration: 200 * time.Millisecond,
		}
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		_, err := RateLimit(ctx, params)
		require.ErrorIs(t, err, context.Canceled)
	})
}
