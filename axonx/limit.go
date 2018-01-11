package axonx

import "time"

type Limiter interface {
	Init(limit int) Limiter
	Acquire()
	Release()
	Stop()
}

type NoLimiter struct {
}

func (n *NoLimiter) Init(limit int) Limiter {
	return n
}

func (n *NoLimiter) Acquire() {
	return
}

func (n *NoLimiter) Release() {

}

func (n *NoLimiter) Stop() {
}

func NewNoLimiter() Limiter {
	return &NoLimiter{}
}

type PoolLimit struct {
	poolSize   int
	bufferChan chan struct{}
}

func (p *PoolLimit) Init(limit int) Limiter {
	p.poolSize = limit
	p.bufferChan = make(chan struct{}, limit)
	return p
}

func (p *PoolLimit) Acquire() {
	p.bufferChan <- struct{}{}
}

func (p *PoolLimit) Release() {
	<-p.bufferChan
}

func (p *PoolLimit) Stop() {
	p.poolSize = 0
	close(p.bufferChan)
}

func NewPoolLimit(limit int) Limiter {
	return (&PoolLimit{}).Init(limit)
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

func (t *TimeLimiter) Acquire() {
	<-t.LimitChan
}

func (t *TimeLimiter) Release() {
	return
}

func (t *TimeLimiter) Stop() {
	close(t.LimitChan)
	t.Tick.Stop()
}
