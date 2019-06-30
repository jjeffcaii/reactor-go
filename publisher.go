package rs

import "context"

type Publisher interface {
  Subscribe(context.Context, Subscriber) Disposable
}
