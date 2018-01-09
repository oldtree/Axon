package axonx

import "time"

type Limiter interface {
	Init(limit int) Limiter
	Limit()
	Stop()
}

type NoLimiter struct {
}

func (n *NoLimiter) Init(limit int) Limiter {
	return n
}

func (n *NoLimiter) Limit() {
	return
}

func (n *NoLimiter) Stop() {

}

func NewNoLimiter() Limiter {
	return &NoLimiter{}
}

type TimeLimiter struct {
	StartTime time.Time
	Tick      *time.Ticker
	LimitChan chan struct{}
}

func NewTimeLimiter(limit int) Limiter {
	return (&TimeLimiter{}).Init(limit)
}

func (t *TimeLimiter) Init(limit int) Limiter {
	t.Tick = time.NewTicker(time.Second / time.Duration(limit))
	t.StartTime = time.Now()
	t.LimitChan = make(chan struct{}, 1)
	go func() {
		for _ = range t.Tick.C {
			t.LimitChan <- struct{}{}
		}
	}()
	return t
}

func (t *TimeLimiter) Limit() {
	<-t.LimitChan
}

func (t *TimeLimiter) Stop() {
	close(t.LimitChan)
	t.Tick.Stop()
}
