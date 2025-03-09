package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

const (
	statCancel   = -1
	statError    = -2
	statComplete = 2
)

// Alias
type (
	Disposable = reactor.Disposable
)

//type FlatMapper func(value reactor.Any) Mono
//type Combinator func(values ...*reactor.Item) (reactor.Any, error)

// Map transforms the item emitted by this Mono by applying a synchronous function to it.
//Map(reactor.Transformer) Mono
// FlatMap transforms the item emitted by this Mono asynchronously, returning the value emitted by another Mono.
//FlatMap(FlatMapper) Mono
// ZipWith combines the result from this mono and another into a tuple.Tuple.
//ZipWith(other Mono) Mono
// ZipCombineWith combines the result from this mono and another, you can customize the Combinator.
//ZipCombineWith(other Mono, cmb Combinator) Mono

// Mono is a Reactive Streams Publisher with basic rx operators that completes successfully by emitting an element, or with an error.
type Mono[T any] interface {
	reactor.Publisher[T]
	// Filter tests the result and replay it if predicate returns true. Otherwise complete without value.
	Filter(reactor.Predicate[T]) Mono[T]
	// SubscribeOn set the scheduler.Scheduler when reactor.Subscriber subscribes this Mono.
	SubscribeOn(scheduler.Scheduler) Mono[T]
	// Block subscribes to this Mono and block indefinitely until a next signal is received.
	// Returns that value/error, or nil if the Mono completes empty.
	Block(context.Context) (T, error)
	// DoOnNext adds a behavior triggered when the Mono emits a data successfully.
	DoOnNext(reactor.FnOnNext[T]) Mono[T]
	// DoOnComplete adds a behavior triggered when the Mono completes successfully (includes empty).
	DoOnComplete(reactor.FnOnComplete) Mono[T]
	// DoOnSubscribe adds a behavior (side effect) triggered when the Mono is done being subscribed, that is to say when a Subscription has been produced by the Publisher and passed to the Subscriber.OnSubscribe(Subscription).
	DoOnSubscribe(reactor.FnOnSubscribe) Mono[T]
	// DoOnError adds a behavior triggered when the Mono completes with an error.
	DoOnError(reactor.FnOnError) Mono[T]
	// DoOnCancel adds a behavior triggered when the Mono is cancelled.
	DoOnCancel(reactor.FnOnCancel) Mono[T]
	// DoFinally adds a behavior triggering after the Mono terminates for any reason, including cancellation.
	DoFinally(reactor.FnOnFinally) Mono[T]
	// DoOnDiscard description.
	DoOnDiscard(reactor.FnOnDiscard[T]) Mono[T]
	// SwitchIfEmpty fallbacks to an alternative Mono if this mono is completed without data.
	SwitchIfEmpty(alternative Mono[T]) Mono[T]
	// SwitchIfError fallbacks to an alternative Mono if this mono is completed without an error.
	SwitchIfError(alternativeFunc func(error) Mono[T]) Mono[T]
	// SwitchValueIfError fallbacks to an alternative value if this mono is completed without an error.
	SwitchValueIfError(v T) Mono[T]
	// DelayElement delays this Mono element (Subscriber.OnNext signal) by a given duration.
	DelayElement(delay time.Duration) Mono[T]
	// Timeout propagates an Error in case no item arrives within the given Duration.
	Timeout(timeout time.Duration) Mono[T]
	// Raw returns an internal RawPublisher.
	Raw() reactor.RawPublisher[T]
}
