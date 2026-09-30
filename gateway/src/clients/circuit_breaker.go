package clients

import (
	"sync"
	"time"
)

type State int

const (
	STATE_CLOSED State = iota
	STATE_HALF_OPEN
	STATE_OPEN

	DEFAULT_N            int           = 5
	DEFAULT_THRESHOLD    float32       = 0.4
	DEFAULT_REFRESH_TIME time.Duration = 10 * time.Second
)

type countCircuitBreaker struct {
	cycle_failures_arr []bool
	cycle_idx          int
	failure_count      int
	failure_threshold  int
	state              State
	half_open_delta    time.Duration
	last_closed        time.Time
	mt                 sync.Mutex
}

type CircuitBreakerStrategy interface {
	OnRequestCompleted(success bool)
	GetState() State
}

func GetCountCircuitBreaker(n int, threshold float32, refresh_time time.Duration) CircuitBreakerStrategy {
	return &countCircuitBreaker{
		make([]bool, n),
		0,
		0,
		int(float32(n) * threshold),
		STATE_CLOSED,
		refresh_time,
		time.Time{},
		sync.Mutex{},
	}
}

func DefaultCircuitBreaker() CircuitBreakerStrategy {
	return GetCountCircuitBreaker(DEFAULT_N, DEFAULT_THRESHOLD, DEFAULT_REFRESH_TIME)
}

func (cnt *countCircuitBreaker) OnRequestCompleted(success bool) {
	cnt.mt.Lock()
	defer cnt.mt.Unlock()

	if cnt.cycle_failures_arr[cnt.cycle_idx] {
		cnt.failure_count--
	}
	cnt.cycle_failures_arr[cnt.cycle_idx] = !success
	if !success {
		cnt.failure_count++
	}
	cnt.cycle_idx = (cnt.cycle_idx + 1) % len(cnt.cycle_failures_arr)
	cnt.check_on_half_open()
	if success && cnt.state == STATE_HALF_OPEN {
		cnt.close()
	} else if cnt.failure_count > cnt.failure_threshold {
		cnt.open()
	}
}

func IsOpen(cx CircuitBreakerStrategy) bool {
	return cx.GetState() == STATE_OPEN
}

func (cnt *countCircuitBreaker) open() {
	if cnt.state == STATE_CLOSED || cnt.state == STATE_HALF_OPEN {
		cnt.state = STATE_OPEN
		cnt.last_closed = time.Now().Add(cnt.half_open_delta)
	}
}

func (cnt *countCircuitBreaker) check_on_half_open() {
	if cnt.state == STATE_OPEN {
		if time.Now().After(cnt.last_closed) {
			cnt.state = STATE_HALF_OPEN
		}
	}
}

func (cnt *countCircuitBreaker) close() {
	if cnt.state == STATE_HALF_OPEN {
		cnt.state = STATE_CLOSED
		clear(cnt.cycle_failures_arr)
		cnt.failure_count = 0
		cnt.cycle_idx = 0
	}
}

func (cnt *countCircuitBreaker) GetState() State {
	cnt.mt.Lock()
	defer cnt.mt.Unlock()

	cnt.check_on_half_open()
	return cnt.state
}
