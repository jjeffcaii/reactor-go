package reactor

// Processor combines the Publisher and Subscriber.
type Processor interface {
	Publisher
	Subscriber
}
