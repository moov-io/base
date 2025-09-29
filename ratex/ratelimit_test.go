package ratex

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExecRetryable(t *testing.T) {
	ctx := context.Background()

	t.Run("Success on first try", func(t *testing.T) {
		closure := func(ctx context.Context) (string, error) {
			return "success", nil
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxRetries:  3,
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
			if attempts < 2 {
				attempts++
				return "", errors.New("retryable error")
			}
			return "success", nil
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxRetries:  3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.NoErrorf(t, err, "Expected success, got error: %v", err)
		require.Equalf(t, "success", result, "Expected result 'success', got: %v", result)
	})

	t.Run("Non-retryable failure", func(t *testing.T) {
		closure := func(ctx context.Context) (string, error) {
			return "", errors.New("non-retryable error")
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return false },
			MaxRetries:  3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.Errorf(t, err, "Expected non-retryable error, got: %v", err)
		require.Empty(t, result)
		require.Equal(t, "non-retryable error", err.Error())
	})

	t.Run("Retryable failures exceeding MaxRetries", func(t *testing.T) {
		closure := func(ctx context.Context) (string, error) {
			return "", errors.New("retryable error")
		}
		params := RetryParams{
			ShouldRetry: func(err error) bool { return true },
			MaxRetries:  3,
			MinDuration: 10 * time.Millisecond,
			MaxDuration: 50 * time.Millisecond,
		}
		result, err := ExecRetryable(ctx, closure, params)
		require.Errorf(t, err, "Expected error after exceeding max retries, got: %v", err)
		require.Empty(t, result)
		require.Equal(t, "hit max tries 3: try 3 of 3: retryable error", err.Error())
	})
}
