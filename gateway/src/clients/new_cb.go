package clients

import (
	"sync"
	"time"
)

type timeBasedCircuitBreaker struct {
	cycle_failures_arr []bool
	expiration         []time.Time
	cycle_idx          int
	failure_count      int
	failure_threshold  int
	state              State
	half_open_delta    time.Duration
	last_closed        time.Time
	mt                 sync.Mutex
}

func GetNewCountCircuitBreaker(n int, threshold float32, refresh_time time.Duration) CircuitBreakerStrategy {
	return &timeBasedCircuitBreaker{
		make([]bool, n),
		make([]time.Time, n),
		0,
		0,
		int(float32(n) * threshold),
		STATE_CLOSED,
		refresh_time,
		time.Time{},
		sync.Mutex{},
	}
}

func (cnt *timeBasedCircuitBreaker) OnRequestCompleted(success bool) {
	cnt.mt.Lock()
	defer cnt.mt.Unlock()

	if cnt.cycle_failures_arr[cnt.cycle_idx] {
		cnt.failure_count--
	}
	cnt.cycle_failures_arr[cnt.cycle_idx] = !success
	cnt.expiration[cnt.cycle_idx] = time.Now().Add(cnt.half_open_delta)
	if !success {
		cnt.failure_count++
	}
	cnt.cycle_idx = (cnt.cycle_idx + 1) % len(cnt.cycle_failures_arr)
	cnt.clearOrphans()

	if cnt.state == STATE_HALF_OPEN {
		if success {
			cnt.close()
		} else {
			cnt.open()
		}
	} else if cnt.failure_count > cnt.failure_threshold {
		cnt.open()
	}
}

func (cnt *timeBasedCircuitBreaker) GetState() State {
	cnt.mt.Lock()
	defer cnt.mt.Unlock()

	cnt.check_on_half_open()
	return cnt.state
}

func (cnt *timeBasedCircuitBreaker) clearOrphans() {
	now := time.Now()
	for i, failed := range cnt.cycle_failures_arr {
		if failed && cnt.expiration[i].Before(now) {
			cnt.cycle_failures_arr[i] = false
			cnt.failure_count--
		}
	}
}

func (cnt *timeBasedCircuitBreaker) open() {
	if cnt.state == STATE_CLOSED || cnt.state == STATE_HALF_OPEN {
		cnt.state = STATE_OPEN
		cnt.last_closed = time.Now().Add(cnt.half_open_delta)
	}
}

func (cnt *timeBasedCircuitBreaker) check_on_half_open() {
	if cnt.state == STATE_OPEN {
		cnt.clearOrphans()
		if time.Now().After(cnt.last_closed) && cnt.failure_count <= cnt.failure_threshold {
			cnt.state = STATE_HALF_OPEN
		}
	}
}

func (cnt *timeBasedCircuitBreaker) close() {
	if cnt.state == STATE_HALF_OPEN {
		cnt.state = STATE_CLOSED
		clear(cnt.cycle_failures_arr)
		cnt.failure_count = 0
		cnt.cycle_idx = 0
	}
}
