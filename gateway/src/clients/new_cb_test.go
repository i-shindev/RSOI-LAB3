package clients

import (
	"testing"
	"testing/synctest"
	"time"
)

// n=5, threshold=0.4 -> failure_threshold=2, CB opens on the 3rd live failure.
// Each result lives for `window`, after that it must "disappear".
const window = 3 * time.Second

func newTestCB() CircuitBreakerStrategy {
	return GetNewCountCircuitBreaker(5, 0.4, window)
}

func report(cb CircuitBreakerStrategy, results ...bool) {
	for _, r := range results {
		cb.OnRequestCompleted(r)
	}
}

func expectState(t *testing.T, cb CircuitBreakerStrategy, want State) {
	t.Helper()
	if got := cb.GetState(); got != want {
		t.Fatalf("state = %v, want %v", got, want)
	}
}

func TestTimeCB_OpensAfterThreeFailures(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false, false)
		expectState(t, cb, STATE_CLOSED)
		report(cb, false)
		expectState(t, cb, STATE_OPEN)
	})
}

func TestTimeCB_ExpiredFailuresAreNotCounted(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false, false)
		time.Sleep(window + time.Second)
		report(cb, false, false) // only 2 live failures
		expectState(t, cb, STATE_CLOSED)
	})
}

func TestTimeCB_SlidingWindow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false) // t=0, expires at t=3
		time.Sleep(2 * time.Second)
		report(cb, false) // t=2
		time.Sleep(1500 * time.Millisecond)
		report(cb, false) // t=3.5, first one already gone
		expectState(t, cb, STATE_CLOSED)
		time.Sleep(500 * time.Millisecond)
		report(cb, false) // t=4: failures at 2, 3.5, 4
		expectState(t, cb, STATE_OPEN)
	})
}

func TestTimeCB_OpenBecomesHalfOpenWhenFailuresExpire(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false, false, false)
		expectState(t, cb, STATE_OPEN)
		time.Sleep(window - time.Second)
		expectState(t, cb, STATE_OPEN)
		time.Sleep(2 * time.Second)
		expectState(t, cb, STATE_HALF_OPEN)
	})
}

// Whole ring buffer filled with failures, all of them expire.
func TestTimeCB_FullBufferExpires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false, false, false, false, false)
		expectState(t, cb, STATE_OPEN)
		time.Sleep(window + time.Second)
		expectState(t, cb, STATE_HALF_OPEN)
	})
}

// Buffer wrapped more than once: newer entries overwrite older ones before they expire.
func TestTimeCB_OverwrittenEntriesExpire(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, true, true, true, true, true, true, false, false) // 8 writes into 5 slots
		time.Sleep(window + time.Second)
		report(cb, false, false) // the old 2 failures are gone -> 2 live
		expectState(t, cb, STATE_CLOSED)
	})
}

// A failure that already expired must not be subtracted again when its slot is overwritten.
func TestTimeCB_NoDoubleDecrement(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false, false)
		time.Sleep(window + time.Second)
		report(cb, true, true, true, true, true) // wraps over the expired failures
		report(cb, false, false, false)
		expectState(t, cb, STATE_OPEN)
	})
}

func TestTimeCB_HalfOpenSuccessCloses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false, false, false)
		time.Sleep(window + time.Second)
		expectState(t, cb, STATE_HALF_OPEN)
		report(cb, true)
		expectState(t, cb, STATE_CLOSED)
		// after closing the CB must behave like a fresh one
		report(cb, false, false)
		expectState(t, cb, STATE_CLOSED)
		report(cb, false)
		expectState(t, cb, STATE_OPEN)
	})
}

func TestTimeCB_HalfOpenFailureReopens(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cb := newTestCB()
		report(cb, false, false, false)
		time.Sleep(window + time.Second)
		expectState(t, cb, STATE_HALF_OPEN)
		report(cb, false)
		expectState(t, cb, STATE_OPEN)
	})
}
