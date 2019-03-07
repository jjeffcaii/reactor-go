package rs

import "context"

type Signal int8

const (
	SignalDefault Signal = iota
	SignalComplete
	SignalCancel
	SignalError
)

type Publisher interface {
	Subscribe(ctx context.Context, first OpSubscriber, others ...OpSubscriber)
}

type Subscriber interface {
	OnSubscribe(s Subscription)
	OnNext(v interface{})
	OnComplete()
	OnError(err error)
}

type Subscription interface {
	Request(n int)
	Cancel()
}

type Processor interface {
	Publisher
	Subscriber
}

type Producer interface {
	Next(v interface{})
	Error(e error)
	Complete()
}

type Scheduler interface {
	Do(ctx context.Context, fn func(ctx context.Context))
}

type Flux interface {
}
