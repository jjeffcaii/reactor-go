package flux

import (
	"context"
	"log"

	reactor "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"go.uber.org/atomic"
)

type fluxPublishOn struct {
	source    reactor.RawPublisher
	sc        scheduler.Scheduler
	prefetch  int
	lowTide   int
	queueSize int
}

func newFluxPublishOn(source reactor.RawPublisher, sc scheduler.Scheduler, prefetch, lowTide, queueSize int) *fluxPublishOn {
	return &fluxPublishOn{
		source:    source,
		sc:        sc,
		prefetch:  prefetch,
		lowTide:   lowTide,
		queueSize: queueSize,
	}
}

func (f *fluxPublishOn) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	w := f.sc.Worker()
	actual := internal.NewCoreSubscriber(ctx, &publishOnSubscriber{
		actual:    internal.ExtractRawSubscriber(sub),
		prefetch:  f.prefetch,
		lowTide:   f.lowTide,
		worker:    w,
		ch:        make(chan interface{}, f.lowTide),
		stat:      atomic.NewInt32(0),
		doing:     atomic.NewBool(false),
		requested: atomic.NewInt64(0),
		fetched:   atomic.NewInt64(0),
		consumed:  atomic.NewInt64(0),
	})
	f.source.SubscribeWith(ctx, actual)
}

type publishOnSubscriber struct {
	actual    reactor.Subscriber
	prefetch  int
	lowTide   int
	worker    scheduler.Worker
	su        reactor.Subscription
	ch        chan interface{}
	stat      *atomic.Int32
	doing     *atomic.Bool
	requested *atomic.Int64
	fetched   *atomic.Int64
	consumed  *atomic.Int64
}

func (p *publishOnSubscriber) Cancel() {
	if p.stat.CAS(0, statCancel) {
		p.su.Cancel()
	}
}

func (p *publishOnSubscriber) Request(n int) {
	req := n
	if n == reactor.RequestInfinite {
		req = p.lowTide
	}
	p.requested.Add(int64(req))
	p.trySchedule()
}

func (p *publishOnSubscriber) OnComplete() {
	if p.stat.CAS(0, statComplete) {
		p.actual.OnComplete()
	}
}

func (p *publishOnSubscriber) OnError(e error) {
	if p.stat.CAS(0, statError) {
		p.actual.OnError(e)
	}
}

func (p *publishOnSubscriber) OnNext(v interface{}) {
	p.ch <- v
	p.trySchedule()
}

func (p *publishOnSubscriber) OnSubscribe(su reactor.Subscription) {
	p.su = su
	p.actual.OnSubscribe(p)
	p.fetchMore(p.prefetch)
}

func (p *publishOnSubscriber) fetchMore(n int) {
	p.fetched.Add(int64(n))
	log.Println("fetch n:", n)
	p.su.Request(n)
}

func (p *publishOnSubscriber) trySchedule() {
	if !p.doing.CAS(false, true) {
		return
	}
	p.worker.Do(func() {
		var reqMore bool
		for {
			if p.stat.Load() != 0 {
				// game over
				break
			}
			requested, consumed := p.requested.Load(), p.consumed.Load()
			if requested <= consumed {
				// request more
				reqMore = true
				break
			}
			fetched := p.fetched.Load()
			if consumed >= fetched {
				p.fetchMore(p.lowTide)
			}
			next, ok := <-p.ch
			if !ok {
				break
			}
			p.actual.OnNext(next)
			p.consumed.Inc()
		}
		if reqMore {
			p.doing.Store(false)
			p.Request(p.lowTide)
		}
	})
}
