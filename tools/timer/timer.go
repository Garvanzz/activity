package timer

import "time"

type Timer struct {
	cb       func(time.Time, int64)
	interval time.Duration
	closed   bool
}

func New(interval time.Duration, cb func(time.Time, int64)) *Timer {
	return &Timer{
		cb:       cb,
		interval: interval,
	}
}

func (t *Timer) Begin() {
	t.run()
}

func (t *Timer) Stop() {
	t.closed = true
}

func (t *Timer) run() {
	now := time.Now()
	t.cb(now, 0)

	if !t.closed {
		time.AfterFunc(t.interval, t.run)
	}
}
