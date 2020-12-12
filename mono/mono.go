package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

// Alias
type (
	Any        = reactor.Any
	Disposable = reactor.Disposable
)

type FlatMapper func(reactor.Any) Mono
type Combinator func(values ...*reactor.Item) (reactor.Any, error)

// Mono is a Reactive Streams Publisher with basic rx operators that completes successfully by emitting an element, or with an error.
type Mono interface {
	reactor.Publisher
	// Filter tests the result and replay it if predicate returns true. Otherwise complete without value.
	Filter(reactor.Predicate) Mono
	// Map transforms the item emitted by this Mono by applying a synchronous function to it.
	Map(reactor.Transformer) Mono
	// FlatMap transforms the item emitted by this Mono asynchronously, returning the value emitted by another Mono.
	FlatMap(FlatMapper) Mono
	// SubscribeOn set the scheduler.Scheduler when reactor.Subscriber subscribes this Mono.
	SubscribeOn(scheduler.Scheduler) Mono
	// Block subscribes to this Mono and block indefinitely until a next signal is received.
	// Returns that value/error, or nil if the Mono completes empty.
	Block(context.Context) (Any, error)
	// DoOnNext adds a behavior triggered when the Mono emits a data successfully.
	DoOnNext(reactor.FnOnNext) Mono
	// DoOnComplete adds a behavior triggered when the Mono completes successfully (includes empty).
	DoOnComplete(reactor.FnOnComplete) Mono
	// DoOnSubscribe adds a behavior (side-effect) triggered when the Mono is done being subscribed, that is to say when a Subscription has been produced by the Publisher and passed to the Subscriber.OnSubscribe(Subscription).
	DoOnSubscribe(reactor.FnOnSubscribe) Mono
	// DoOnError adds a behavior triggered when the Mono completes with an error.
	DoOnError(reactor.FnOnError) Mono
	// DoOnCancel adds a behavior triggered when the Mono is cancelled.
	DoOnCancel(reactor.FnOnCancel) Mono
	// DoFinally adds a behavior triggering after the Mono terminates for any reason, including cancellation.
	DoFinally(reactor.FnOnFinally) Mono
	// DoOnDiscard description.
	DoOnDiscard(reactor.FnOnDiscard) Mono
	// SwitchIfEmpty fallbacks to an alternative Mono if this mono is completed without data.
	SwitchIfEmpty(alternative Mono) Mono
	// SwitchIfError fallbacks to an alternative Mono if this mono is completed without an error.
	SwitchIfError(alternativeFunc func(error) Mono) Mono
	// SwitchValueIfError fallbacks to an alternative value if this mono is completed without an error.
	SwitchValueIfError(v Any) Mono
	// DelayElement delays this Mono element (Subscriber.OnNext signal) by a given duration.
	DelayElement(delay time.Duration) Mono
	// Timeout propagates a Error in case no item arrives within the given Duration.
	Timeout(timeout time.Duration) Mono
	// ZipWith combines the result from this mono and another into a tuple.Tuple.
	ZipWith(other Mono) Mono
	// ZipCombineWith combines the result from this mono and another, you can customize the Combinator.
	ZipCombineWith(other Mono, cmb Combinator) Mono
	// Raw returns an internal RawPublisher.
	Raw() reactor.RawPublisher
}
