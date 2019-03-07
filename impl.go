package rs

import "context"

type unicastProcessor struct {
	n            int32
	c            func(Producer)
	q            *Queue
	e            error
	sub          Subscriber
	sig          Signal
	pubScheduler Scheduler
	subScheduler Scheduler
}

func (p *unicastProcessor) Request(n int) {
	if n < 1 {
		panic("invalid N")
	}
	p.n = int32(n)
}

func (p *unicastProcessor) Cancel() {
	p.sig = SignalCancel
	_ = p.q.Close()
}

func (p *unicastProcessor) Next(v interface{}) {
	if p.sig == SignalDefault {
		_ = p.q.Add(v)
	}
}

func (p *unicastProcessor) Error(e error) {
	p.e = e
	p.sig = SignalError
	_ = p.q.Close()
}

func (p *unicastProcessor) Complete() {
	p.sig = SignalComplete
	_ = p.q.Close()
}

func (p *unicastProcessor) start() {

}

func (p *unicastProcessor) Subscribe(ctx context.Context, first OpSubscriber, others ...OpSubscriber) {
	sub := &subscriberBuilder{
		m: make(map[opSubType][]interface{}, 8),
	}
	first(sub)
	for _, it := range others {
		it(sub)
	}
	p.sub = sub
	p.pubScheduler.Do(ctx, func(ctx context.Context) {
		p.c(p)
	})
	p.subScheduler.Do(ctx, func(ctx context.Context) {
		p.OnSubscribe(p)
		for {
			v, ok := p.q.Poll(ctx, func(ctx context.Context, q *Queue) {
				q.RequestN(p.n)
			})
			if !ok {
				break
			}
			p.OnNext(v)
		}
		switch p.sig {
		case SignalComplete:
			p.OnComplete()
		case SignalError:
			p.OnError(p.e)
		case SignalCancel:
			//TODO
		}
	})
}

func (p *unicastProcessor) OnSubscribe(s Subscription) {
	p.sub.OnSubscribe(s)
}

func (p *unicastProcessor) OnNext(v interface{}) {
	p.sub.OnNext(v)
}

func (p *unicastProcessor) OnComplete() {
	p.sub.OnComplete()
}

func (p *unicastProcessor) OnError(err error) {
	p.sub.OnError(err)
}

func NewFlux(fn func(producer Producer)) Publisher {
	return &unicastProcessor{
		c:            fn,
		q:            NewQueue(16),
		n:            256,
		pubScheduler: Elastic(),
		subScheduler: Immet(),
	}
}
