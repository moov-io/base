package ratex

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"

	"github.com/moov-io/base/telemetry"
)

type RateLimitParams struct {
	RateLimiter *rate.Limiter // can be nil to create a new rate limiter
	TryCount    int           `otel:"try_count"`
	MinDuration time.Duration `otel:"min_duration_n"`
	MaxDuration time.Duration `otel:"max_duration_ns"`
}

func RateLimit(ctx context.Context, params RateLimitParams) (*rate.Limiter, error) {
	ctx, span := telemetry.StartSpan(ctx, "rate-limiter-wait",
		trace.WithAttributes(telemetry.StructAttributes(params)...))
	defer span.End()

	var (
		err error
	)

	params.RateLimiter, err = generateRateLimiter(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("generating rate limiter: %w", err)
	}

	err = params.RateLimiter.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	return params.RateLimiter, nil
}

// generateRateLimiter initializes a new rate limiter or sets a new limit on it.
func generateRateLimiter(ctx context.Context, params RateLimitParams) (*rate.Limiter, error) {
	rateLimitDuration, err := generateRateLimitDuration(params.TryCount, params.MinDuration, params.MaxDuration)
	if err != nil {
		return nil, fmt.Errorf("generating rate limit duration: %w", err)
	}

	rateLimitInterval := rate.Every(rateLimitDuration)
	if params.RateLimiter == nil {
		params.RateLimiter = rate.NewLimiter(rateLimitInterval, 1)
		// A rate limiter is initialized with 1 token. So the first call to Wait will not wait/block, only subsequent calls to Wait will.
		// Call wait immediately after initializing to use up token and ensure we trigger a delay on next call to Wait.
		if err := params.RateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter wait: %w", err)
		}
	} else {
		params.RateLimiter.SetLimit(rateLimitInterval)
	}

	return params.RateLimiter, nil
}

// generateRateLimitDuration returns a random value between min-max duration multiplied by the multiplier.
func generateRateLimitDuration(multiplier int, minDuration, maxDuration time.Duration) (time.Duration, error) {
	minVal := minDuration.Milliseconds()
	maxVal := maxDuration.Milliseconds()

	maxRand, err := rand.Int(rand.Reader, big.NewInt(maxVal-minVal))
	if err != nil {
		return 0, fmt.Errorf("rand int: %w", err)
	}
	waitInterval := (minVal + maxRand.Int64()) * int64(multiplier)
	return time.Millisecond * time.Duration(waitInterval), nil
}

type RetryParams struct {
	ShouldRetry func(err error) bool
	MaxRetries  int           `otel:"max_retries"`
	MinDuration time.Duration `otel:"min_duration_n"`
	MaxDuration time.Duration `otel:"max_duration_ns"`
}

func ExecRetryable[R any](ctx context.Context, closure func(ctx context.Context) (R, error), params RetryParams) (R, error) {
	var (
		rateLimiter *rate.Limiter
		retVal      R
		err         error
	)

	tryFunc := func(ctx context.Context, tryCount int) (R, error) {
		tryCtx, span := telemetry.StartSpan(ctx, "try",
			trace.WithAttributes(
				attribute.Int("try_count", tryCount),
				attribute.Int("max_tries", params.MaxRetries),
			),
		)
		defer span.End()
		return closure(tryCtx)
	}

	for i := range params.MaxRetries {
		tryCount := i + 1
		retVal, err = tryFunc(ctx, tryCount)

		// no error means success - break out
		if err == nil {
			break
		}

		// if the error doesn't have one of the flags do not retry, instead return the error
		if !params.ShouldRetry(err) {
			return retVal, err
		}

		// record event if we'll be attempting retries
		err = fmt.Errorf("try %d of %d: %w", tryCount, params.MaxRetries, err)
		telemetry.AddEvent(ctx, err.Error())

		if tryCount != params.MaxRetries {
			// If error and we haven't hit max tries,
			// generate rate limiter to delay retries.
			// This will jitter a wait time before the next iteration.
			//
			// We continue on rate limit errors and retry without waiting
			params := RateLimitParams{
				RateLimiter: rateLimiter,
				TryCount:    tryCount,
				MinDuration: params.MinDuration,
				MaxDuration: params.MaxDuration,
			}
			rateLimiter, err = RateLimit(ctx, params)
			if err != nil {
				telemetry.AddEvent(ctx, fmt.Sprintf("rate limit: %s", err.Error()))
				continue
			}
		}
	}

	if err != nil {
		return retVal, fmt.Errorf("hit max tries %d: %w", params.MaxRetries, err)
	}
	return retVal, nil
}
