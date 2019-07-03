package rs

import "math"

const RequestInfinite = math.MaxInt32

type Subscription interface {
  Request(n int)
  Cancel()
}

type Subscriber interface {
  OnComplete()
  OnError(error)
  OnNext(Subscription, interface{})
  OnSubscribe(Subscription)
}

type subscriber struct {
  fnOnSubscribe FnOnSubscribe
  fnOnNext      FnOnNext
  fnOnComplete  FnOnComplete
  fnOnError     FnOnError
}

func (p *subscriber) OnComplete() {
  if p.fnOnComplete != nil {
    p.fnOnComplete()
  }
}

func (p *subscriber) OnError(err error) {
  if p.fnOnError != nil {
    p.fnOnError(err)
  }
}

func (p *subscriber) OnSubscribe(s Subscription) {
  if p.fnOnSubscribe != nil {
    p.fnOnSubscribe(s)
  } else {
    s.Request(RequestInfinite)
  }
}

func (p *subscriber) OnNext(s Subscription, v interface{}) {
  if p.fnOnNext != nil {
    p.fnOnNext(s, v)
  }
}

type SubscriberOption func(*subscriber)

func OnNext(onNext FnOnNext) SubscriberOption {
  return func(s *subscriber) {
    s.fnOnNext = onNext
  }
}

func OnComplete(onComplete FnOnComplete) SubscriberOption {
  return func(s *subscriber) {
    s.fnOnComplete = onComplete
  }
}

func OnError(onError FnOnError) SubscriberOption {
  return func(i *subscriber) {
    i.fnOnError = onError
  }
}

func OnSubscribe(onSubscribe FnOnSubscribe) SubscriberOption {
  return func(i *subscriber) {
    i.fnOnSubscribe = onSubscribe
  }
}

func NewSubscriber(opts ...SubscriberOption) Subscriber {
  s := &subscriber{}
  for _, opt := range opts {
    opt(s)
  }
  return s
}
