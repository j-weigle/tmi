package tmi

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiterWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestRateLimiterWait, reason:short mode")
	}
	rl := NewRateLimiter(RLimJoinDefault)
	// RLimJoinDefault has burst of 20, and rate of 0.5 tokens / second
	// therefore 10 tokens over should result in 5 seconds total of waiting
	want := float64(time.Second * 5)
	waitCalls := rl.burst + 10

	start := time.Now()
	for i := 0; i < waitCalls; i++ {
		rl.Wait()
	}
	end := time.Now()

	diff := float64(end.Sub(start))

	// error margin of 0.025 seconds
	var epsilon = 0.025 * float64(time.Second)
	gtMin := diff > (want - epsilon)
	ltMax := diff < (want + epsilon)
	if !(gtMin && ltMax) {
		t.Errorf("got %v, want %v", diff, want)
	}
}

func TestMultiRoutineRateLimiterWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestMultiRoutineRateLimiterWait, reason:short mode")
	}
	rl := NewRateLimiter(RLimJoinDefault)
	// RLimJoinDefault refill rate of 0.5 tokens / second
	// therefore 20 tokens over should result in 10 seconds total of waiting
	want := float64(time.Second * 10)
	gR1WaitCalls := (rl.burst + 20) / 2
	gR2WaitCalls := gR1WaitCalls

	wg := &sync.WaitGroup{}
	wg.Add(2)

	start := time.Now()
	go func() {
		for i := 0; i < gR1WaitCalls; i++ {
			rl.Wait()
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < gR2WaitCalls; i++ {
			rl.Wait()
		}
		wg.Done()
	}()

	wg.Wait()
	end := time.Now()

	diff := float64(end.Sub(start))

	// error margin of 0.05 seconds
	var epsilon = 0.05 * float64(time.Second)
	gtMin := diff > (want - epsilon)
	ltMax := diff < (want + epsilon)
	if !(gtMin && ltMax) {
		t.Errorf("got %v, want %v", diff, want)
	}
}
