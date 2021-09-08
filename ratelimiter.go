package tmi

import (
	"sync"
	"time"
)

// RateLimit is used for creating a new RateLimiter
type RateLimit struct {
	// Total number of uses allowed before rate limiting kicks in.
	Burst int
	// The refill rate of the RateLimiter. Example for 20 per 10 seconds: time.Second * 10 / 20
	Rate time.Duration
}

// RateLimiter is a token bucket.
type RateLimiter struct {
	mu     sync.Mutex
	tokens float64   // current number of tokens held
	burst  int       // total number of uses when full before being rate limited
	rate   float64   // refill rate in tokens per second
	last   time.Time // last time tokens were updated
}

// burst presets all set to half the rate limit as a precaution
var (
	// RLimJoinDefault is the regular account rate limit 20 attempts per 10s
	RLimJoinDefault = RateLimit{Burst: 10, Rate: time.Second * 10 / 20}

	// RLimJoinVerified is the verified account rate limit 2000 attempts per 10s
	RLimJoinVerified = RateLimit{Burst: 1000, Rate: time.Second * 10 / 2000}

	// RLimMsgDefault is the regular account rate limit 20 messages per 30s
	RLimMsgDefault = RateLimit{Burst: 10, Rate: time.Second * 30 / 20}

	// RLimMsgMod is the mod/broadcaster/VIP account rate limit 100 messages per 30s
	RLimMsgMod = RateLimit{Burst: 50, Rate: time.Second * 30 / 100}

	// RLimGlobalDefault is the verified account rate limit 7500 global messages per 30s
	// assumed to be global limit for other accounts as well since it is undocumented
	RLimGlobalDefault = RateLimit{Burst: 3750, Rate: time.Second * 30 / 7500}

	// RLimWhisperDefault is the rate limit for any account of 100 messages per 60s
	// 100 / minute is more constricting than 3 / second, so it is chosen
	RLimWhisperDefault = RateLimit{Burst: 2, Rate: time.Minute / 100}
)

// NewRateLimiter returns a new RateLimiter based on the RateLimit provided.
// It initializes the number of tokens available to RateLimit.Burst.
func NewRateLimiter(rl RateLimit) *RateLimiter {
	return &RateLimiter{
		tokens: float64(rl.Burst),
		burst:  rl.Burst,
		rate:   1 / float64(rl.Rate.Seconds()),
	}
}

// Wait refills tokens based on time passed since last call, then claims a token.
// If tokens are negative after taking one, the number of tokens below zero is divided by
// the refill rate to determine how long to wait before a token becomes available.
// Wait is thread safe.
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()

	rl.replenish()

	var wait time.Duration
	rl.tokens -= 1
	if rl.tokens < 0 {
		wait = time.Duration((-rl.tokens / rl.rate) * float64(time.Second))
	}

	rl.mu.Unlock()

	if wait > 0 {
		t := time.NewTimer(wait)
		<-t.C
	}
}

// replenish calculates how many tokens to replenish based on the time difference
// between the last time tokens were replenished and now, and sets the total to the
// replenished amount plus the current amount (up to RateLimiter.burst).
// replenish requires that the mutex lock is held for the RateLimiter.
func (rl *RateLimiter) replenish() {
	var now = time.Now()
	var elapsed = now.Sub(rl.last)

	var delta float64 = elapsed.Seconds() * rl.rate
	var tokens float64 = rl.tokens + delta

	var burst = float64(rl.burst)
	if tokens > burst {
		tokens = burst
	}
	rl.tokens = tokens
	rl.last = now
}
