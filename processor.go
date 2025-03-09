package reactor

// Processor combines the Publisher and Subscriber.
type Processor[T any] interface {
	Publisher[T]
	Subscriber[T]
}
